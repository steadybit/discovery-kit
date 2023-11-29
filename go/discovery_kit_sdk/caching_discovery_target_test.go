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

func Test_target_caching(t *testing.T) {
	ctx := context.Background()

	discovery := newMockTargetDiscovery()
	cached := NewCachedTargetDiscovery(discovery, WithRefreshTargetsNow())

	discovery.WaitForNextDiscovery()
	first, _ := cached.DiscoverTargets(ctx)
	second, _ := cached.DiscoverTargets(ctx)

	assert.Equal(t, first, second)

	discovery.AssertNumberOfCalls(t, "DiscoverTargets", 1)
}

func Test_target_timeout(t *testing.T) {
	ctx := context.Background()

	discovery := newMockTargetDiscovery()
	cached := NewCachedTargetDiscovery(discovery, WithTargetsRefreshTimeout(1*time.Second))

	discovery.On("DiscoverTargets", mock.Anything).Unset()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{}}, nil).Once()
	call := discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{}, nil).Once()
	call.RunFn = func(args mock.Arguments) {
		time.Sleep(2 * time.Second)
	}
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{}}, nil).Once()

	cached.Refresh(ctx)
	first, _ := cached.DiscoverTargets(ctx)
	assert.Len(t, first, 1)

	cached.Refresh(ctx)
	_, secondErr := cached.DiscoverTargets(ctx)
	assert.ErrorIs(t, secondErr, ErrDiscoveryTimeout)

	cached.Refresh(ctx)
	third, _ := cached.DiscoverTargets(ctx)
	assert.Len(t, third, 1)
}

func Test_target_caching_error(t *testing.T) {
	ctx := context.Background()

	discovery := newMockTargetDiscovery()
	ch := make(chan struct{})
	cached := NewCachedTargetDiscovery(discovery, WithRefreshTargetsTrigger(ctx, ch, 0))

	discovery.On("DiscoverTargets", mock.Anything).Unset()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{}}, nil).Once()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{}, errors.New("test")).Once()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{}}, nil).Once()

	ch <- struct{}{}
	discovery.WaitForNextDiscovery()
	data, err := cached.DiscoverTargets(ctx)
	assert.NoError(t, err)
	assert.Len(t, data, 1)

	ch <- struct{}{}
	discovery.WaitForNextDiscovery()
	data, err = cached.DiscoverTargets(ctx)
	assert.Error(t, err)
	assert.Len(t, data, 0)

	ch <- struct{}{}
	discovery.WaitForNextDiscovery()
	data, err = cached.DiscoverTargets(ctx)
	assert.NoError(t, err)
	assert.Len(t, data, 1)
}

func Test_target_caching_should_recover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery := newMockTargetDiscovery()
	cached := NewCachedTargetDiscovery(discovery, WithRefreshTargetsInterval(ctx, 20*time.Millisecond))

	recovered := discovery_kit_api.Target{
		Id:         "recovered",
		TargetType: "example",
		Label:      "Example Target",
		Attributes: map[string][]string{},
	}

	discovery.On("DiscoverTargets", ctx).Unset()
	discovery.On("DiscoverTargets", ctx).Panic("test").Once()
	discovery.On("DiscoverTargets", ctx).Return([]discovery_kit_api.Target{recovered}, nil)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		targets, _ := cached.DiscoverTargets(ctx)
		assert.Contains(c, targets, recovered)
	}, 1000*time.Millisecond, 10*time.Millisecond)
}

func Test_target_cache_interval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	discovery := newMockTargetDiscovery()
	cached := NewCachedTargetDiscovery(discovery, WithRefreshTargetsInterval(ctx, 20*time.Millisecond))

	//should cache
	discovery.WaitForNextDiscovery()
	first, _ := cached.DiscoverTargets(ctx)
	second, _ := cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first, _ = cached.DiscoverTargets(ctx)
	discovery.WaitForNextDiscovery()
	second, _ = cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)
	discovery.WaitForNextDiscovery()
	third, _ := cached.DiscoverTargets(ctx)
	assert.NotEqual(t, second, third)

	//should not refresh cache after cancel
	cancel()
	first, _ = cached.DiscoverTargets(ctx)
	time.Sleep(200 * time.Millisecond)
	second, _ = cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)
}

func Test_target_cache_trigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	discovery := newMockTargetDiscovery()
	ch := make(chan struct{})
	cached := NewCachedTargetDiscovery(discovery, WithRefreshTargetsTrigger(ctx, ch, 0))

	//should cache
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	first, _ := cached.DiscoverTargets(ctx)
	second, _ := cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first, _ = cached.DiscoverTargets(ctx)
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	second, _ = cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)

	//should not refresh cache after cancel
	cancel()
	first, _ = cached.DiscoverTargets(ctx)
	time.Sleep(200 * time.Millisecond)
	second, _ = cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)
}

func Test_target_cache_trigger_throttle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery := newMockTargetDiscovery()
	ch := make(chan struct{})
	cached := NewCachedTargetDiscovery(discovery, WithRefreshTargetsTrigger(ctx, ch, 500*time.Millisecond))

	//should refresh cache
	first, _ := cached.DiscoverTargets(ctx)
	ch <- struct{}{}
	ch <- struct{}{}
	ch <- struct{}{}
	ch <- struct{}{}
	discovery.WaitForNextDiscovery(func() {
		ch <- struct{}{}
	})
	second, _ := cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)
	discovery.AssertNumberOfCalls(t, "DiscoverTargets", 2)
}

func Test_target_cache_update(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}

	discovery := newMockTargetDiscovery()
	ch := make(chan string)
	updateFn := func(data []discovery_kit_api.Target, update string) ([]discovery_kit_api.Target, error) {
		defer wg.Done()
		if update == "clear" {
			return []discovery_kit_api.Target{}, nil
		}
		return data, nil
	}
	cached := NewCachedTargetDiscovery(discovery,
		WithRefreshTargetsNow(),
		WithTargetsUpdate(ctx, ch, updateFn),
	)

	//should cache
	discovery.WaitForNextDiscovery()
	first, _ := cached.DiscoverTargets(ctx)
	second, _ := cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)

	//should update cache
	first, _ = cached.DiscoverTargets(ctx)
	wg.Add(1)
	go func() {
		ch <- "clear"
	}()
	wg.Wait()
	second, _ = cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)
	assert.Empty(t, second)
}
