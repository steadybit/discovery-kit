// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func Test_enrichmentData_caching(t *testing.T) {
	ctx := context.Background()

	discovery := newMockEnrichmentDataDiscovery()
	cached := CachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataNow())

	discovery.WaitForNextDiscovery()
	first := cached.DiscoverEnrichmentData(ctx)
	second := cached.DiscoverEnrichmentData(ctx)

	assert.Equal(t, first, second)

	discovery.AssertNumberOfCalls(t, "DiscoverEnrichmentData", 1)
}

func Test_enrichmentData_caching_shoud_recover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery := newMockEnrichmentDataDiscovery()
	cached := CachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataInterval(ctx, 20*time.Millisecond))

	discovery.On("DiscoverEnrichmentData", ctx).Unset()
	discovery.On("DiscoverEnrichmentData", ctx).Panic("test").Once()
	discovery.On("DiscoverEnrichmentData", ctx).Return([]discovery_kit_api.EnrichmentData{
		{
			Id:                 "recovered",
			EnrichmentDataType: "example",
			Attributes:         map[string][]string{},
		},
	})

	assert.Eventually(t, func() bool {
		data := cached.DiscoverEnrichmentData(ctx)
		return len(data) > 0 && data[0].Id == "recovered"
	}, 1000*time.Millisecond, 10*time.Millisecond)
}

func Test_enrichmentData_cache_interval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	discovery := newMockEnrichmentDataDiscovery()
	cached := CachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataInterval(ctx, 20*time.Millisecond))

	//should cache
	discovery.WaitForNextDiscovery()
	first := cached.DiscoverEnrichmentData(ctx)
	second := cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first = cached.DiscoverEnrichmentData(ctx)
	discovery.WaitForNextDiscovery()
	second = cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)
	discovery.WaitForNextDiscovery()
	third := cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, second, third)

	//should not refresh cache after cancel
	cancel()
	first = cached.DiscoverEnrichmentData(ctx)
	time.Sleep(200 * time.Millisecond)
	second = cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)
}

func Test_enrichmentData_cache_trigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	discovery := newMockEnrichmentDataDiscovery()
	ch := make(chan struct{})
	cached := CachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataTrigger(ctx, ch))

	//should cache
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	first := cached.DiscoverEnrichmentData(ctx)
	second := cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first = cached.DiscoverEnrichmentData(ctx)
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	second = cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)

	//should not refresh cache after cancel
	cancel()
	first = cached.DiscoverEnrichmentData(ctx)
	time.Sleep(200 * time.Millisecond)
	second = cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)
}

func Test_enrichmentData_cache_update(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}

	discovery := newMockEnrichmentDataDiscovery()
	ch := make(chan string)
	updateFn := func(data []discovery_kit_api.EnrichmentData, update string) []discovery_kit_api.EnrichmentData {
		defer wg.Done()
		if update == "clear" {
			return []discovery_kit_api.EnrichmentData{}
		}
		return data
	}
	cached := CachedEnrichmentDataDiscovery(discovery,
		WithRefreshEnrichmentDataNow(),
		WithEnrichmentDataUpdate(ctx, ch, updateFn),
	)

	//should cache
	discovery.WaitForNextDiscovery()
	first := cached.DiscoverEnrichmentData(ctx)
	second := cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)

	//should update cache
	first = cached.DiscoverEnrichmentData(ctx)
	wg.Add(1)
	go func() {
		ch <- "clear"
	}()
	wg.Wait()
	second = cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)
	assert.Empty(t, second)
}
