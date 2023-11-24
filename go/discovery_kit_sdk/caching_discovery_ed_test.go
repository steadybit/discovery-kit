// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"errors"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

func Test_enrichmentData_caching(t *testing.T) {
	ctx := context.Background()

	discovery := newMockEnrichmentDataDiscovery()
	cached := NewCachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataNow())

	discovery.WaitForNextDiscovery()
	first, _ := cached.DiscoverEnrichmentData(ctx)
	second, _ := cached.DiscoverEnrichmentData(ctx)

	assert.Equal(t, first, second)

	discovery.AssertNumberOfCalls(t, "DiscoverEnrichmentData", 1)
}

func Test_enrichmentData_caching_error(t *testing.T) {
	ctx := context.Background()

	discovery := newMockEnrichmentDataDiscovery()
	ch := make(chan struct{})
	cached := NewCachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataTrigger(ctx, ch, 0))

	discovery.On("DiscoverEnrichmentData", mock.Anything).Unset()
	discovery.On("DiscoverEnrichmentData", mock.Anything).Return([]discovery_kit_api.EnrichmentData{{}}, nil).Once()
	discovery.On("DiscoverEnrichmentData", mock.Anything).Return([]discovery_kit_api.EnrichmentData{}, errors.New("test")).Once()
	discovery.On("DiscoverEnrichmentData", mock.Anything).Return([]discovery_kit_api.EnrichmentData{{}}, nil).Once()

	ch <- struct{}{}
	discovery.WaitForNextDiscovery()
	data, err := cached.DiscoverEnrichmentData(ctx)
	assert.NoError(t, err)
	assert.Len(t, data, 1)

	ch <- struct{}{}
	discovery.WaitForNextDiscovery()
	data, err = cached.DiscoverEnrichmentData(ctx)
	assert.Error(t, err)
	assert.Len(t, data, 0)

	ch <- struct{}{}
	discovery.WaitForNextDiscovery()
	data, err = cached.DiscoverEnrichmentData(ctx)
	assert.NoError(t, err)
	assert.Len(t, data, 1)
}

func Test_enrichmentData_caching_should_recover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery := newMockEnrichmentDataDiscovery()
	cached := NewCachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataInterval(ctx, 20*time.Millisecond))

	recovered := discovery_kit_api.EnrichmentData{
		Id:                 "recovered",
		EnrichmentDataType: "example",
		Attributes:         map[string][]string{},
	}

	discovery.On("DiscoverEnrichmentData", ctx).Unset()
	discovery.On("DiscoverEnrichmentData", ctx).Panic("test").Once()
	discovery.On("DiscoverEnrichmentData", ctx).Return([]discovery_kit_api.EnrichmentData{recovered}, nil)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		targets, _ := cached.DiscoverEnrichmentData(ctx)
		assert.Contains(c, targets, recovered)
	}, 1000*time.Millisecond, 10*time.Millisecond)
}

func Test_enrichmentData_cache_interval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	discovery := newMockEnrichmentDataDiscovery()
	cached := NewCachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataInterval(ctx, 20*time.Millisecond))

	//should cache
	discovery.WaitForNextDiscovery()
	first, _ := cached.DiscoverEnrichmentData(ctx)
	second, _ := cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first, _ = cached.DiscoverEnrichmentData(ctx)
	discovery.WaitForNextDiscovery()
	second, _ = cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)
	discovery.WaitForNextDiscovery()
	third, _ := cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, second, third)

	//should not refresh cache after cancel
	cancel()
	first, _ = cached.DiscoverEnrichmentData(ctx)
	time.Sleep(200 * time.Millisecond)
	second, _ = cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)
}

func Test_enrichmentData_cache_trigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	discovery := newMockEnrichmentDataDiscovery()
	ch := make(chan struct{})
	cached := NewCachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataTrigger(ctx, ch, 0))

	//should cache
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	first, _ := cached.DiscoverEnrichmentData(ctx)
	second, _ := cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first, _ = cached.DiscoverEnrichmentData(ctx)
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	second, _ = cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)

	//should not refresh cache after cancel
	cancel()
	first, _ = cached.DiscoverEnrichmentData(ctx)
	time.Sleep(200 * time.Millisecond)
	second, _ = cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)
}

func Test_enrichmentData_cache_trigger_debounced(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery := newMockEnrichmentDataDiscovery()
	ch := make(chan struct{})
	cached := NewCachedEnrichmentDataDiscovery(discovery, WithRefreshEnrichmentDataTrigger(ctx, ch, 500*time.Millisecond))

	//should refresh cache
	first, _ := cached.DiscoverEnrichmentData(ctx)
	ch <- struct{}{}
	ch <- struct{}{}
	ch <- struct{}{}
	ch <- struct{}{}
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	second, _ := cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)
	discovery.AssertNumberOfCalls(t, "DiscoverEnrichmentData", 2)
}

func Test_enrichmentData_cache_update(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}

	discovery := newMockEnrichmentDataDiscovery()
	ch := make(chan string)
	updateFn := func(data []discovery_kit_api.EnrichmentData, update string) ([]discovery_kit_api.EnrichmentData, error) {
		defer wg.Done()
		if update == "clear" {
			return []discovery_kit_api.EnrichmentData{}, nil
		}
		return data, nil
	}
	cached := NewCachedEnrichmentDataDiscovery(discovery,
		WithRefreshEnrichmentDataNow(),
		WithEnrichmentDataUpdate(ctx, ch, updateFn),
	)

	//should cache
	discovery.WaitForNextDiscovery()
	first, _ := cached.DiscoverEnrichmentData(ctx)
	second, _ := cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)

	//should update cache
	first, _ = cached.DiscoverEnrichmentData(ctx)
	wg.Add(1)
	go func() {
		ch <- "clear"
	}()
	wg.Wait()
	second, _ = cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)
	assert.Empty(t, second)
}
