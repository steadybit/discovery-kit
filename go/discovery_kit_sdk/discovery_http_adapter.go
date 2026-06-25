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

// normalizeTargets returns a defensive copy of the target list where each Attributes map is a
// fresh copy with multi-valued slices sorted, and (optionally) the steadybit.group attribute set.
//
// The discovery's underlying maps are never mutated, which keeps concurrent calls to handleDiscover
// safe — without the copy, two requests could write to the same cached map while a third was
// iterating it for JSON encoding, causing a "concurrent map iteration and map write" panic.
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
// into a single self-contained string per entry.
func normalizeTargets(targets []discovery_kit_api.Target, group string) []discovery_kit_api.Target {
	out := make([]discovery_kit_api.Target, len(targets))
	for i, t := range targets {
		out[i] = t
		out[i].Attributes = normalizeAttributes(t.Attributes, group)
	}
	return out
}

func normalizeEnrichmentData(data []discovery_kit_api.EnrichmentData, group string) []discovery_kit_api.EnrichmentData {
	out := make([]discovery_kit_api.EnrichmentData, len(data))
	for i, d := range data {
		out[i] = d
		out[i].Attributes = normalizeAttributes(d.Attributes, group)
	}
	return out
}

func normalizeAttributes(src map[string][]string, group string) map[string][]string {
	extra := 0
	if group != "" {
		extra = 1
	}
	dst := make(map[string][]string, len(src)+extra)
	for k, v := range src {
		if len(v) > 1 {
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
