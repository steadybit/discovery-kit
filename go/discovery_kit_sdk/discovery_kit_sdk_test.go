// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
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
	"net/http"
	"testing"
	"time"
)

type TestCase struct {
	Name string
	Fn   func(*testing.T, *TestContext)
}

type TestContext struct {
	targetDiscovery         *MockTargetDiscovery
	enrichmentDataDiscovery *MockEnrichmentDataDiscovery
	cachedDiscovery         *CachedTargetDiscovery
	refreshTrigger          chan struct{}
	api                     client.DiscoveryAPI
	r                       *resty.Client
}

func Test_discoveryHttpAdapter(t *testing.T) {
	extlogging.InitZeroLog()
	clearDefaultServeMux()

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
		{
			Name: "test enrichment rules",
			Fn:   testEDEnrichmentRules,
		},
		{
			Name: "test cached target discovery",
			Fn:   testCachedDiscovery,
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

	trigger := make(chan struct{})
	cachedDiscovery := NewCachedTargetDiscovery(targetDiscovery,
		WithRefreshTargetsNow(),
		WithRefreshTargetsTrigger(context.Background(), trigger, 0),
	)

	serverPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	go func() {
		Register(edDiscovery)
		Register(cachedDiscovery)

		exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(GetDiscoveryList))
		exthttp.Listen(exthttp.ListenOpts{Port: serverPort})
	}()

	httpClient := resty.New().SetBaseURL(fmt.Sprintf("http://localhost:%d", serverPort))
	api := client.NewDiscoveryClient("", httpClient)

	tc := &TestContext{
		targetDiscovery:         targetDiscovery,
		cachedDiscovery:         cachedDiscovery,
		enrichmentDataDiscovery: edDiscovery,
		api:                     api,
		r:                       httpClient,
		refreshTrigger:          trigger,
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Fn(t, tc)
		})
	}
}

func testCachedDiscovery(t *testing.T, tc *TestContext) {
	res, err := tc.r.R().Get("/example/discovery/discovered-targets")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	etag := res.Header().Get("ETag")
	assert.NotEmpty(t, etag)

	res, err = tc.r.R().SetHeader("If-None-Match", etag).Get("/example/discovery/discovered-targets")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotModified, res.StatusCode())

	tc.refreshTrigger <- struct{}{}
	tc.targetDiscovery.WaitForNextDiscovery()
	res, err = tc.r.R().SetHeader("If-None-Match", etag).Get("/example/discovery/discovered-targets")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
}

func testValidateEndpoints(t *testing.T, tc *TestContext) {
	assert.NoError(t, validate.ValidateEndpointReferences("/", tc.r))
}

func testTargetDiscovery(t *testing.T, tc *TestContext) {
	targets, err := tc.api.DiscoverTargets("example")
	require.NoError(t, err)
	assert.Len(t, targets, 1)
	assert.Equal(t, "target", targets[0].Id)

	targetType, err := tc.api.DescribeTargetForId("example")
	require.NoError(t, err)
	assert.Equal(t, "example", targetType.Id)

	attributes, err := tc.api.DescribeAllAttributes()
	require.NoError(t, err)
	assert.Contains(t, attributes.Attributes, discovery_kit_api.AttributeDescription{
		Attribute: "target.created",
		Label: discovery_kit_api.PluralLabel{
			One:   "Creation Date",
			Other: "Creation Dates",
		},
	})
}

func testEDDiscovery(t *testing.T, tc *TestContext) {
	eds, err := tc.api.DiscoverEnrichmentData("example-ed")
	require.NoError(t, err)
	assert.Len(t, eds, 1)
	assert.Equal(t, "example-ed", eds[0].Id)

	attributes, err := tc.api.DescribeAllAttributes()
	require.NoError(t, err)
	assert.Contains(t, attributes.Attributes, discovery_kit_api.AttributeDescription{
		Attribute: "example-ed.created",
		Label: discovery_kit_api.PluralLabel{
			One:   "Creation Date",
			Other: "Creation Dates",
		},
	})
}

func testEDEnrichmentRules(t *testing.T, tc *TestContext) {
	list, err := tc.api.ListDiscoveries()
	require.NoError(t, err)
	assert.Len(t, list.TargetEnrichmentRules, 2)

	ruleIds := make([]string, 0, len(list.TargetEnrichmentRules))
	for _, ruleRef := range list.TargetEnrichmentRules {
		rule, err := tc.api.DescribeEnrichmentRule(ruleRef)
		require.NoError(t, err)
		ruleIds = append(ruleIds, rule.Id)
	}

	assert.Contains(t, ruleIds, "enrichmentRule-1")
	assert.Contains(t, ruleIds, "enrichmentRule-2")
}

func Test_unwrap(t *testing.T) {
	clearDefaultServeMux()

	Register(NewCachedTargetDiscovery(newMockTargetDiscovery()))
	Register(NewCachedEnrichmentDataDiscovery(newMockEnrichmentDataDiscovery()))

	discoveries := GetDiscoveryList()

	assert.Len(t, discoveries.Discoveries, 2)
	assert.Len(t, discoveries.TargetAttributes, 1)
	assert.Len(t, discoveries.TargetTypes, 1)
	assert.Len(t, discoveries.TargetEnrichmentRules, 2)
}

func clearDefaultServeMux() {
	http.DefaultServeMux = new(http.ServeMux)
}
