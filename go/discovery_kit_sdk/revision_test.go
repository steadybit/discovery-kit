// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package discovery_kit_sdk

import (
	"testing"
	"time"

	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/stretchr/testify/assert"
)

func TestRegisterBumpsRevision(t *testing.T) {
	ClearRegisteredDiscoveries()
	t.Cleanup(ClearRegisteredDiscoveries)

	// Use a discovery id unique to this test: registration installs routes on the process-global
	// http.DefaultServeMux, and ClearRegisteredDiscoveries does not remove them, so reusing an id
	// another test already registered would panic on a duplicate route.
	discovery := &MockTargetDiscovery{MockDiscovery{Now: time.Now}}
	discovery.On("Describe").Return(discovery_kit_api.DiscoveryDescription{Id: "revision-test"})
	discovery.On("DescribeTarget").Return(discovery_kit_api.TargetDescription{
		Id:    "revision-test",
		Label: discovery_kit_api.PluralLabel{One: "Revision Test", Other: "Revision Tests"},
	})
	discovery.On("DescribeAttributes").Return([]discovery_kit_api.AttributeDescription{})

	before := exthttp.Revision()
	Register(discovery)
	assert.NotEqual(t, before, exthttp.Revision(), "Register must bump the index revision")
}

func TestClearRegisteredDiscoveriesBumpsRevision(t *testing.T) {
	before := exthttp.Revision()
	ClearRegisteredDiscoveries()
	assert.NotEqual(t, before, exthttp.Revision(), "ClearRegisteredDiscoveries must bump the index revision")
}
