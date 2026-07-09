// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
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
		description.Discover.CallInterval = new(defaultCallInterval)
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
		if group != "" {
			targets = withGroupAttribute(targets, group)
		}
		body.Targets = new(targets)
	}
	if e, ok := a.discovery.(EnrichmentDataDiscovery); ok {
		data, err := e.DiscoverEnrichmentData(newCtx)
		a.checkForDuplicateEnrichmentData(data)
		if err != nil {
			allErrs = errors.Join(allErrs, err)
		}
		if group != "" {
			data = withGroupAttributeEnrichment(data, group)
		}
		body.EnrichmentData = new(data)
	}
	if allErrs != nil {
		exthttp.WriteError(w, extension_kit.ToError("Discovery Failed", allErrs))
		return
	}
	exthttp.WriteBody(w, body)
}

// withGroupAttribute returns a new slice of targets where each target's Attributes map is a
// fresh copy with the group attribute set. The discovery's underlying maps are never mutated,
// which keeps concurrent calls to handleDiscover safe — without this, two requests could
// write to the same cached map while a third was iterating it for JSON encoding, causing a
// "concurrent map iteration and map write" panic.
func withGroupAttribute(targets []discovery_kit_api.Target, group string) []discovery_kit_api.Target {
	out := make([]discovery_kit_api.Target, len(targets))
	for i, t := range targets {
		out[i] = t
		out[i].Attributes = copyAttributesWithGroup(t.Attributes, group)
	}
	return out
}

func withGroupAttributeEnrichment(data []discovery_kit_api.EnrichmentData, group string) []discovery_kit_api.EnrichmentData {
	out := make([]discovery_kit_api.EnrichmentData, len(data))
	for i, d := range data {
		out[i] = d
		out[i].Attributes = copyAttributesWithGroup(d.Attributes, group)
	}
	return out
}

func copyAttributesWithGroup(src map[string][]string, group string) map[string][]string {
	dst := make(map[string][]string, len(src)+1)
	maps.Copy(dst, src)
	dst[groupAttributeKey] = []string{group}
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
