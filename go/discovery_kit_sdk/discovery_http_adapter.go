// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"net/http"
	"strconv"
)

const (
	defaultCallInterval = "30s"
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
	if t, ok := a.discovery.(TargetDiscovery); ok {
		targets, err := t.DiscoverTargets(newCtx)
		a.checkForDuplicateTargets(targets)
		if err != nil {
			allErrs = errors.Join(allErrs, err)
		}
		body.Targets = extutil.Ptr(targets)
	}
	if e, ok := a.discovery.(EnrichmentDataDiscovery); ok {
		data, err := e.DiscoverEnrichmentData(newCtx)
		a.checkForDuplicateEnrichmentData(data)
		if err != nil {
			allErrs = errors.Join(allErrs, err)
		}
		body.EnrichmentData = extutil.Ptr(data)
	}
	if allErrs != nil {
		exthttp.WriteError(w, extension_kit.ToError("Discovery Failed", allErrs))
		return
	}
	exthttp.WriteBody(w, body)
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
