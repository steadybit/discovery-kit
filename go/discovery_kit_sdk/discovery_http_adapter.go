// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
)

const (
	defaultCallInterval = "30s"
	groupAttributeKey   = "steadybit.group"
	groupEnvVar         = "STEADYBIT_EXTENSION_DISCOVERY_GROUP"
)

func newDiscoveryHttpAdapter(discovery Discovery) *discoveryHttpAdapter {
	description := getDescriptionWithDefaults(discovery)
	adapter := &discoveryHttpAdapter{
		description: description,
		discovery:   discovery,
		rootPath:    fmt.Sprintf("/%s/discovery", description.Id),
	}
	return adapter
}

func getDescriptionWithDefaults(discovery Discovery) discovery_kit_api.DiscoveryDescription {
	description := discovery.Describe()
	if description.Discover.Path == "" {
		description.Discover.Path = fmt.Sprintf("/%s/discovery/discovered-targets", description.Id)
	}
	if description.Discover.Method == "" {
		description.Discover.Method = discovery_kit_api.GET
	}
	if description.Discover.CallInterval == nil {
		description.Discover.CallInterval = extutil.Ptr(defaultCallInterval)
	}
	return description
}

type discoveryHttpAdapter struct {
	description discovery_kit_api.DiscoveryDescription
	discovery   Discovery
	rootPath    string
}

func (a discoveryHttpAdapter) registerHandlers() {
	discover := a.handleDiscover
	if m, ok := a.discovery.(p); ok {
		discover = exthttp.IfNoneMatchHandler(func() string {
			return strconv.FormatInt(m.LastModified().UnixMilli(), 10)
		}, discover)
	}
	exthttp.RegisterHttpHandler(a.rootPath, a.handleGetDescription)
	exthttp.RegisterHttpHandler(a.description.Discover.Path, discover)
}

func (a discoveryHttpAdapter) handleGetDescription(w http.ResponseWriter, _ *http.Request, _ []byte) {
	exthttp.WriteBody(w, a.description)
}

type HttpRequestContextKey string

func (a discoveryHttpAdapter) handleDiscover(w http.ResponseWriter, r *http.Request, _ []byte) {
	body := discovery_kit_api.DiscoveryData{}
	var allErrs error
	var key HttpRequestContextKey = "httpRequest"
	newCtx := context.WithValue(r.Context(), key, r)
	group := os.Getenv(groupEnvVar)
	if t, ok := a.discovery.(TargetDiscovery); ok {
		targets, err := t.DiscoverTargets(newCtx)
		a.checkForDuplicateTargets(targets)
		if err != nil {
			allErrs = errors.Join(allErrs, err)
		}
		targets = normalizeTargets(targets, group)
		body.Targets = extutil.Ptr(targets)
	}
	if e, ok := a.discovery.(EnrichmentDataDiscovery); ok {
		data, err := e.DiscoverEnrichmentData(newCtx)
		a.checkForDuplicateEnrichmentData(data)
		if err != nil {
			allErrs = errors.Join(allErrs, err)
		}
		data = normalizeEnrichmentData(data, group)
		body.EnrichmentData = extutil.Ptr(data)
	}
	if allErrs != nil {
		exthttp.WriteError(w, extension_kit.ToError("Discovery Failed", allErrs))
		return
	}
	exthttp.WriteBody(w, body)
}

// normalizeTargets ensures every multi-valued attribute slice in the response is sorted, and
// optionally injects the steadybit.group attribute on every target. It has two branches:
//
//   - Fast path (group == "" AND every multi-valued slice is already sorted): the input slice is
//     returned verbatim, no allocation. The discovery's underlying maps are caller-visible through
//     the return value, so callers MUST NOT mutate them. This keeps the no-work-needed case free
//     of GC pressure on large-fan-out discoveries.
//   - Slow path (group set OR any multi-valued slice unsorted): returns a fresh slice where each
//     Attributes map is a fresh copy with multi-valued slices sorted into a new backing array.
//     The source maps are never mutated, so concurrent calls to handleDiscover stay safe under
//     JSON encoding even though the discovery cache may rebuild its internal maps in-flight.
//
// The sort is the load-bearing part: the platform's target-diff detector compares multi-valued
// attribute slices by position, so extensions that build them from a Go map or a Kubernetes
// client-go lister (both of which iterate in randomized order) would otherwise make every such
// target look "changed" on every discovery cycle, driving needless writes to the platform's
// target store. Sorting at the SDK level keeps every Go-based extension safe by default.
//
// Note for extensions: multi-valued attributes are normalized as **sets**. If you use parallel
// multi-valued attributes (where the i-th element of one pairs with the i-th of another), pre-sort
// your source so the per-attribute sort here keeps the pairing aligned, or encode paired values
// into a single self-contained string per entry. See docs/discovery-api.md, section
// "Multi-valued Attribute Values".
func normalizeTargets(targets []discovery_kit_api.Target, group string) []discovery_kit_api.Target {
	if !targetsNeedNormalize(targets, group) {
		return targets
	}
	out := make([]discovery_kit_api.Target, len(targets))
	for i, t := range targets {
		out[i] = t
		out[i].Attributes = normalizeAttributes(t.Attributes, group)
	}
	return out
}

func normalizeEnrichmentData(data []discovery_kit_api.EnrichmentData, group string) []discovery_kit_api.EnrichmentData {
	if !enrichmentNeedsNormalize(data, group) {
		return data
	}
	out := make([]discovery_kit_api.EnrichmentData, len(data))
	for i, d := range data {
		out[i] = d
		out[i].Attributes = normalizeAttributes(d.Attributes, group)
	}
	return out
}

// targetsNeedNormalize / enrichmentNeedsNormalize keep the no-copy fast path
// for discoveries with no group configured and whose multi-valued attribute
// slices already come out sorted. Without this, every serve allocates a fresh
// map per target whether or not anything would change — measurable cost on
// large-fan-out discoveries (10k+ targets) polled at tight intervals.
func targetsNeedNormalize(targets []discovery_kit_api.Target, group string) bool {
	if group != "" {
		return true
	}
	for _, t := range targets {
		if attributesNeedNormalize(t.Attributes) {
			return true
		}
	}
	return false
}

func enrichmentNeedsNormalize(data []discovery_kit_api.EnrichmentData, group string) bool {
	if group != "" {
		return true
	}
	for _, d := range data {
		if attributesNeedNormalize(d.Attributes) {
			return true
		}
	}
	return false
}

func attributesNeedNormalize(src map[string][]string) bool {
	for _, v := range src {
		if len(v) > 1 && !sort.StringsAreSorted(v) {
			return true
		}
	}
	return false
}

func normalizeAttributes(src map[string][]string, group string) map[string][]string {
	extra := 0
	if group != "" {
		extra = 1
	}
	dst := make(map[string][]string, len(src)+extra)
	for k, v := range src {
		if len(v) > 1 && !sort.StringsAreSorted(v) {
			sorted := make([]string, len(v))
			copy(sorted, v)
			sort.Strings(sorted)
			dst[k] = sorted
		} else {
			dst[k] = v
		}
	}
	if group != "" {
		dst[groupAttributeKey] = []string{group}
	}
	return dst
}

type duplicateCheckKey struct {
	Id   string
	Type string
}

func (a discoveryHttpAdapter) checkForDuplicateTargets(targets []discovery_kit_api.Target) {
	seenTargets := make(map[duplicateCheckKey]struct{})
	for _, target := range targets {
		key := duplicateCheckKey{Id: target.Id, Type: target.TargetType}
		if _, exists := seenTargets[key]; exists {
			log.Warn().
				Str("id", target.Id).
				Str("targetType", target.TargetType).
				Msg("Duplicate target detected.")
		} else {
			seenTargets[key] = struct{}{}
		}
	}
}

func (a discoveryHttpAdapter) checkForDuplicateEnrichmentData(targets []discovery_kit_api.EnrichmentData) {
	seenTargets := make(map[duplicateCheckKey]struct{})
	for _, target := range targets {
		key := duplicateCheckKey{Id: target.Id, Type: target.EnrichmentDataType}
		if _, exists := seenTargets[key]; exists {
			log.Warn().
				Str("id", target.Id).
				Str("targetType", target.EnrichmentDataType).
				Msg("Duplicate enrichmentData detected.")
		} else {
			seenTargets[key] = struct{}{}
		}
	}
}
