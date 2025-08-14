// Copyright 2025 steadybit GmbH. All rights reserved.

package discovery_kit_sdk

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func Test_target_caching(t *testing.T) {
	ctx := context.Background()

	discovery := newMockTargetDiscovery()
	cached := NewCachedTargetDiscovery(discovery)

	cached.Refresh(context.Background())
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

	trigger := func() {
		ch <- struct{}{}
	}

	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)
	data, err := cached.DiscoverTargets(ctx)
	assert.NoError(t, err)
	assert.Len(t, data, 1)

	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)
	data, err = cached.DiscoverTargets(ctx)
	assert.Error(t, err)
	assert.Len(t, data, 0)

	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)
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
	waitForLastModifiedChanges(t, &cached.CachedDiscovery)
	first, _ := cached.DiscoverTargets(ctx)
	second, _ := cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first, _ = cached.DiscoverTargets(ctx)
	waitForLastModifiedChanges(t, &cached.CachedDiscovery)
	second, _ = cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)
	waitForLastModifiedChanges(t, &cached.CachedDiscovery)
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

	trigger := func() {
		ch <- struct{}{}
	}

	//should cache
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)
	first, _ := cached.DiscoverTargets(ctx)
	second, _ := cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first, _ = cached.DiscoverTargets(ctx)
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)
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
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, func() {
		ch <- struct{}{}
	})
	second, _ := cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)
	discovery.AssertNumberOfCalls(t, "DiscoverTargets", 2)
}

func Test_target_cache_update(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery := newMockTargetDiscovery()
	ch := make(chan string)
	updateFn := func(data []discovery_kit_api.Target, update string) ([]discovery_kit_api.Target, error) {
		if update == "clear" {
			return []discovery_kit_api.Target{}, nil
		}
		return data, nil
	}
	cached := NewCachedTargetDiscovery(discovery,
		WithTargetsUpdate(ctx, ch, updateFn),
	)

	//should cache
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, func() {
		cached.Refresh(context.Background())
	})
	first, _ := cached.DiscoverTargets(ctx)
	second, _ := cached.DiscoverTargets(ctx)
	assert.Equal(t, first, second)

	//should update cache
	first, _ = cached.DiscoverTargets(ctx)
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, func() {
		ch <- "clear"
	})
	second, _ = cached.DiscoverTargets(ctx)
	assert.NotEqual(t, first, second)
	assert.Empty(t, second)
}

func Test_target_string_interning(t *testing.T) {
	largeString := "ID: this is a very large string which should get unshared"
	ctx := context.Background()

	discovery := newMockTargetDiscovery()
	cached := NewCachedTargetDiscovery(discovery)

	discovery.On("DiscoverTargets", mock.Anything).Unset()
	discovery.On("DiscoverTargets", ctx).Return([]discovery_kit_api.Target{{
		Id:    largeString[:2],
		Label: largeString[:2],
		Attributes: map[string][]string{
			largeString[:2]: {largeString[4:]},
		},
	}}, nil)
	cached.Refresh(ctx)
	data, _ := cached.DiscoverTargets(ctx)

	assert.Equal(t, "ID", data[0].Id)
	assert.Equal(t, []string{"this is a very large string which should get unshared"}, data[0].Attributes["ID"])

	assertSliceNotShared(t, largeString, data[0].Id)
	assertSliceNotShared(t, largeString, data[0].Label)
	for _, datum := range data {
		for key, values := range datum.Attributes {
			assertSliceNotShared(t, largeString, key)
			for _, value := range values {
				assertSliceNotShared(t, largeString, value)
			}
		}
	}
}

func Test_target_string_interning_concurrent_modification(t *testing.T) {
	ctx := context.Background()

	discovery := newMockTargetDiscovery()
	cached := NewCachedTargetDiscovery(discovery)

	targets := make([]discovery_kit_api.Target, 1000)
	for i := range targets {
		targets[i] = discovery_kit_api.Target{
			Id:         fmt.Sprintf("target-%d", i),
			TargetType: "example",
			Label:      "Example Target",
			Attributes: map[string][]string{
				"example": {"yes"},
				"id":      {fmt.Sprintf("target-%d", i)},
			},
		}
	}

	discovery.On("DiscoverTargets", mock.Anything).Unset()
	discovery.On("DiscoverTargets", ctx).Return(targets, nil)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Any("err", err).Msgf("recovered")
			}
		}()

		for {
			for i := range targets {
				targets[i].Attributes["loop"] = []string{fmt.Sprintf("loop-%d", i)}
			}
		}
	}()

	assert.NotPanics(t, func() {
		cached.Refresh(ctx)
		_, _ = cached.DiscoverTargets(ctx)
	})
}
