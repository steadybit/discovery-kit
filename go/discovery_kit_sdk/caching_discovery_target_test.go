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

func Test_target_caching(t *testing.T) {
	ctx := context.Background()

	discovery := newMockTargetDiscovery()
	cached := CachedTargetDiscovery(discovery, WithRefreshTargetsNow())

	discovery.WaitForNextDiscovery()
	first := cached.DiscoverTargets(ctx)
	second := cached.DiscoverTargets(ctx)

	assert.Equal(t, first, second)

	discovery.AssertNumberOfCalls(t, "DiscoverTargets", 1)
}

func Test_target_caching_shoud_recover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery := newMockTargetDiscovery()
	cached := CachedTargetDiscovery(discovery, WithRefreshTargetsInterval(ctx, 20*time.Millisecond))

	discovery.On("DiscoverTargets", ctx).Unset()
	discovery.On("DiscoverTargets", ctx).Panic("test").Once()
	discovery.On("DiscoverTargets", ctx).Return([]discovery_kit_api.Target{
		{
			Id:         "recovered",
			TargetType: "example",
			Label:      "Example Target",
			Attributes: map[string][]string{},
		},
	})

	assert.Eventually(t, func() bool {
		targets := cached.DiscoverTargets(ctx)
		return len(targets) > 0 && targets[0].Id == "recovered"
	}, 1000*time.Millisecond, 10*time.Millisecond)
}

func Test_target_cache_interval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	discovery := newMockTargetDiscovery()
	cached := CachedTargetDiscovery(discovery, WithRefreshTargetsInterval(ctx, 20*time.Millisecond))

	//should cache
	discovery.WaitForNextDiscovery()
	first := cached.DiscoverTargets(ctx)
	second := cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first = cached.DiscoverTargets(ctx)
	discovery.WaitForNextDiscovery()
	second = cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)
	discovery.WaitForNextDiscovery()
	third := cached.DiscoverTargets(ctx)
	assert.NotEqual(t, second, third)

	//should not refresh cache after cancel
	cancel()
	first = cached.DiscoverTargets(ctx)
	time.Sleep(200 * time.Millisecond)
	second = cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)
}

func Test_target_cache_trigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	discovery := newMockTargetDiscovery()
	ch := make(chan struct{})
	cached := CachedTargetDiscovery(discovery, WithRefreshTargetsTrigger(ctx, ch))

	//should cache
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	first := cached.DiscoverTargets(ctx)
	second := cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first = cached.DiscoverTargets(ctx)
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	second = cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)

	//should not refresh cache after cancel
	cancel()
	first = cached.DiscoverTargets(ctx)
	time.Sleep(200 * time.Millisecond)
	second = cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)
}

func Test_target_cache_update(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}

	discovery := newMockTargetDiscovery()
	ch := make(chan string)
	updateFn := func(data []discovery_kit_api.Target, update string) []discovery_kit_api.Target {
		defer wg.Done()
		if update == "clear" {
			return []discovery_kit_api.Target{}
		}
		return data
	}
	cached := CachedTargetDiscovery(discovery,
		WithRefreshTargetsNow(),
		WithTargetsUpdate(ctx, ch, updateFn),
	)

	//should cache
	discovery.WaitForNextDiscovery()
	first := cached.DiscoverTargets(ctx)
	second := cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)

	//should update cache
	first = cached.DiscoverTargets(ctx)
	wg.Add(1)
	go func() {
		ch <- "clear"
	}()
	wg.Wait()
	second = cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)
	assert.Empty(t, second)
}
