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
	cachedDiscovery         *CachingTargetDiscovery
	refreshTrigger          chan struct{}
	api                     client.DiscoveryAPI
	r                       *resty.Client
}

func Test_discoveryHttpAdapter(t *testing.T) {
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
	cachedDiscovery := CachedTargetDiscovery(targetDiscovery,
		WithRefreshTargetsNow(),
		WithRefreshTargetsTrigger(context.Background(), trigger),
	)

	serverPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	go func() {
		extlogging.InitZeroLog()
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

func Test_unwrap(t *testing.T) {
	clearDefaultServeMux()

	Register(CachedTargetDiscovery(newMockTargetDiscovery()))
	Register(CachedEnrichmentDataDiscovery(newMockEnrichmentDataDiscovery()))

	discoveries := GetDiscoveryList()

	assert.Len(t, discoveries.Discoveries, 2)
	assert.Len(t, discoveries.TargetAttributes, 1)
	assert.Len(t, discoveries.TargetTypes, 1)
	assert.Len(t, discoveries.TargetEnrichmentRules, 1)
}

func clearDefaultServeMux() {
	http.DefaultServeMux = new(http.ServeMux)
}
