// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/phayes/freeport"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_test/client"
	"github.com/steadybit/discovery-kit/go/discovery_kit_test/validate"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type TestCase struct {
	Name string
	Fn   func(*testing.T, client.DiscoveryAPI, *resty.Client)
}

func Test_discoveryHttpAdapter(t *testing.T) {
	testCases := []TestCase{
		{
			Name: "validate endpoints",
			Fn:   testValidateEndpoints,
		},
		{
			Name: "test target discovery",
			Fn:   testTargetDiscovery,
		},
		{
			Name: "test enrichment data discovery",
			Fn:   testEDDiscovery,
		},
	}

	targetDiscovery := newMockTargetDiscovery()
	targetDiscovery.Now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	edDiscovery := newMockEnrichmentDataDiscovery()
	edDiscovery.Now = func() time.Time {
		return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	serverPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	go func() {
		extlogging.InitZeroLog()
		for _, d := range []interface{}{edDiscovery, targetDiscovery} {
			Register(d)
		}
		exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(GetDiscoveryList))
		exthttp.Listen(exthttp.ListenOpts{Port: serverPort})
	}()

	httpClient := resty.New().SetBaseURL(fmt.Sprintf("http://localhost:%d", serverPort))
	api := client.NewDiscoveryClient("", httpClient)

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Fn(t, api, httpClient)
		})
	}
}

func testValidateEndpoints(t *testing.T, _ client.DiscoveryAPI, r *resty.Client) {
	assert.NoError(t, validate.ValidateEndpointReferences("/", r))
}

func testTargetDiscovery(t *testing.T, api client.DiscoveryAPI, _ *resty.Client) {
	targets, err := api.DiscoverTargets("example")
	require.NoError(t, err)
	assert.Len(t, targets, 1)
	assert.Equal(t, "target", targets[0].Id)

	targetType, err := api.DescribeTargetForId("example")
	require.NoError(t, err)
	assert.Equal(t, "example", targetType.Id)

	attributes, err := api.DescribeAllAttributes()
	require.NoError(t, err)
	assert.Contains(t, attributes.Attributes, discovery_kit_api.AttributeDescription{
		Attribute: "target.created",
		Label: discovery_kit_api.PluralLabel{
			One:   "Creation Date",
			Other: "Creation Dates",
		},
	})
}

func testEDDiscovery(t *testing.T, api client.DiscoveryAPI, _ *resty.Client) {
	eds, err := api.DiscoverEnrichmentData("example-ed")
	require.NoError(t, err)
	assert.Len(t, eds, 1)
	assert.Equal(t, "example-ed", eds[0].Id)

	attributes, err := api.DescribeAllAttributes()
	require.NoError(t, err)
	assert.Contains(t, attributes.Attributes, discovery_kit_api.AttributeDescription{
		Attribute: "example-ed.created",
		Label: discovery_kit_api.PluralLabel{
			One:   "Creation Date",
			Other: "Creation Dates",
		},
	})
}

func Test_unwrap(t *testing.T) {
	Register(CachedTargetDiscovery(newMockTargetDiscovery()))
	Register(CachedEnrichmentDataDiscovery(newMockEnrichmentDataDiscovery()))

	discoveries := GetDiscoveryList()

	assert.Len(t, discoveries.Discoveries, 2)
	assert.Len(t, discoveries.TargetAttributes, 1)
	assert.Len(t, discoveries.TargetTypes, 1)
	assert.Len(t, discoveries.TargetEnrichmentRules, 1)
}
