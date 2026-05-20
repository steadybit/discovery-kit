// Copyright 2026 steadybit GmbH. All rights reserved.

package discovery_kit_sdk

import (
	"context"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
)

func Test_target_group_attribute_injected_on_refresh(t *testing.T) {
	t.Setenv(groupEnvVar, "prod-eu")

	discovery := newMockTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Unset()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{
		{Id: "a", TargetType: "x", Label: "a", Attributes: map[string][]string{"k": {"v"}}},
	}, nil)

	cached := NewCachedTargetDiscovery(discovery)
	cached.Refresh(context.Background())

	got, err := cached.DiscoverTargets(context.Background())
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, []string{"prod-eu"}, got[0].Attributes[groupAttributeKey])
	assert.Equal(t, []string{"v"}, got[0].Attributes["k"])
}

func Test_enrichment_data_group_attribute_injected_on_refresh(t *testing.T) {
	t.Setenv(groupEnvVar, "prod-eu")

	discovery := newMockEnrichmentDataDiscovery()
	discovery.On("DiscoverEnrichmentData", mock.Anything).Unset()
	discovery.On("DiscoverEnrichmentData", mock.Anything).Return([]discovery_kit_api.EnrichmentData{
		{Id: "a", EnrichmentDataType: "x", Attributes: map[string][]string{"k": {"v"}}},
	}, nil)

	cached := NewCachedEnrichmentDataDiscovery(discovery)
	cached.Refresh(context.Background())

	got, err := cached.DiscoverEnrichmentData(context.Background())
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, []string{"prod-eu"}, got[0].Attributes[groupAttributeKey])
}

func Test_no_group_env_means_no_group_attribute(t *testing.T) {
	// no t.Setenv — group env var is unset
	discovery := newMockTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Unset()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{
		{Id: "a", TargetType: "x", Attributes: map[string][]string{"k": {"v"}}},
	}, nil)

	cached := NewCachedTargetDiscovery(discovery)
	cached.Refresh(context.Background())

	got, _ := cached.DiscoverTargets(context.Background())
	_, exists := got[0].Attributes[groupAttributeKey]
	assert.False(t, exists, "group attribute should not be set when env var is empty")
}

// Test_concurrent_reads_are_safe locks in the fix for the
// "concurrent map iteration and map write" panic. Multiple readers iterate the
// cached target attribute maps in parallel while a refresh swaps the slice.
// Run with -race to be meaningful.
func Test_concurrent_reads_are_safe(t *testing.T) {
	t.Setenv(groupEnvVar, "prod-eu")

	discovery := newMockTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Unset()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{
		{Id: "a", TargetType: "x", Attributes: map[string][]string{"k": {"v"}}},
		{Id: "b", TargetType: "x", Attributes: map[string][]string{"k": {"v"}}},
	}, nil)

	cached := NewCachedTargetDiscovery(discovery)
	cached.Refresh(context.Background())

	stop := make(chan struct{})
	var refresher sync.WaitGroup
	refresher.Add(1)
	go func() {
		defer refresher.Done()
		for {
			select {
			case <-stop:
				return
			default:
				cached.Refresh(context.Background())
			}
		}
	}()

	var readers sync.WaitGroup
	for i := 0; i < 4; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for j := 0; j < 200; j++ {
				targets, _ := cached.DiscoverTargets(context.Background())
				for _, tg := range targets {
					for k, v := range tg.Attributes {
						_ = k
						_ = v
					}
				}
			}
		}()
	}
	readers.Wait()
	close(stop)
	refresher.Wait()
}
