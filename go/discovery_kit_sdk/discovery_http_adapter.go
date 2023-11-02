// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"fmt"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"net/http"
	"strconv"
)

const (
	defaultCallInterval = "30s"
	defaultRestrictTo   = discovery_kit_api.LEADER
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
	if description.RestrictTo == nil {
		description.RestrictTo = extutil.Ptr(defaultRestrictTo)
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
	if m, ok := a.discovery.(LastModifiedProvider); ok {
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

func (a discoveryHttpAdapter) handleDiscover(w http.ResponseWriter, r *http.Request, _ []byte) {
	body := discovery_kit_api.DiscoveryData{}
	if t, ok := a.discovery.(TargetDiscovery); ok {
		body.Targets = extutil.Ptr(t.DiscoverTargets(r.Context()))
	}
	if e, ok := a.discovery.(EnrichmentDataDiscovery); ok {
		body.EnrichmentData = extutil.Ptr(e.DiscoverEnrichmentData(r.Context()))
	}
	exthttp.WriteBody(w, body)
}
