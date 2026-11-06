// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errDiscovery = errors.New("api returned 500")

func bareTargetDiscovery() *MockTargetDiscovery {
	discovery := newMockTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Unset()
	return discovery
}

func Test_keep_targets_on_error(t *testing.T) {
	ctx := t.Context()
	discovery := bareTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil).Once()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
	cached := NewCachedTargetDiscovery(discovery, WithKeepTargetsOnError(time.Hour))

	cached.Refresh(ctx)
	lastModified := cached.LastModified()
	cached.Refresh(ctx)

	targets, err := cached.DiscoverTargets(ctx)
	require.NoError(t, err)
	assert.Equal(t, []discovery_kit_api.Target{{Id: "a", Attributes: map[string][]string{}}}, targets)
	assert.Equal(t, lastModified, cached.LastModified())
}

func Test_keep_targets_on_error_reports_error_without_previous_success(t *testing.T) {
	ctx := t.Context()
	discovery := bareTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
	cached := NewCachedTargetDiscovery(discovery, WithKeepTargetsOnError(time.Hour))

	cached.Refresh(ctx)

	_, err := cached.DiscoverTargets(ctx)
	assert.ErrorIs(t, err, errDiscovery)
}

func Test_keep_targets_on_error_reports_error_once_too_old(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		discovery := bareTargetDiscovery()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil).Once()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
		cached := NewCachedTargetDiscovery(discovery, WithKeepTargetsOnError(time.Hour))

		cached.Refresh(ctx)
		time.Sleep(2 * time.Hour)
		cached.Refresh(ctx)

		_, err := cached.DiscoverTargets(ctx)
		assert.ErrorIs(t, err, errDiscovery)
	})
}

func Test_keep_targets_on_error_ignores_failed_updates(t *testing.T) {
	ctx := t.Context()
	discovery := bareTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil)
	cached := NewCachedTargetDiscovery(discovery, WithKeepTargetsOnError(time.Hour))

	cached.Refresh(ctx)
	cached.Update(func([]discovery_kit_api.Target) ([]discovery_kit_api.Target, error) {
		return nil, errDiscovery
	})

	_, err := cached.DiscoverTargets(ctx)
	assert.ErrorIs(t, err, errDiscovery)
}

func Test_retry_targets_on_error_recovers_failed_initial_refresh(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		discovery := bareTargetDiscovery()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery).Once()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil)
		// The pause lets a refresh started by WithRefreshTargetsNow run before the retry option is applied.
		pause := func(*CachedDiscovery[discovery_kit_api.Target]) { time.Sleep(time.Millisecond) }
		cached := NewCachedTargetDiscovery(discovery,
			WithRefreshTargetsNow(),
			pause,
			WithRefreshTargetsInterval(ctx, 3*time.Hour),
			WithRetryTargetsOnError(ctx, 5*time.Minute, 3*time.Hour),
		)

		synctest.Wait()
		_, err := cached.DiscoverTargets(ctx)
		require.ErrorIs(t, err, errDiscovery)

		time.Sleep(5 * time.Minute)
		synctest.Wait()
		targets, err := cached.DiscoverTargets(ctx)
		require.NoError(t, err)
		assert.Len(t, targets, 1)
		discovery.AssertNumberOfCalls(t, "DiscoverTargets", 2)
	})
}

func Test_retry_targets_on_error_backs_off_until_max_delay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		discovery := bareTargetDiscovery()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
		cached := NewCachedTargetDiscovery(discovery, WithRetryTargetsOnError(ctx, 5*time.Minute, 20*time.Minute))

		cached.Refresh(ctx)
		time.Sleep(24 * time.Hour)
		synctest.Wait()

		// Retries after 5, 10 and 20 minutes; 40 minutes would exceed the maximum.
		discovery.AssertNumberOfCalls(t, "DiscoverTargets", 4)
	})
}

func Test_retry_targets_on_error_is_canceled_by_successful_refresh(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		discovery := bareTargetDiscovery()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery).Once()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil)
		cached := NewCachedTargetDiscovery(discovery, WithRetryTargetsOnError(ctx, 5*time.Minute, time.Hour))

		cached.Refresh(ctx)
		cached.Refresh(ctx)
		time.Sleep(time.Hour)
		synctest.Wait()

		discovery.AssertNumberOfCalls(t, "DiscoverTargets", 2)
	})
}

func Test_retry_targets_on_error_does_not_stack_retries(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		discovery := bareTargetDiscovery()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery).Twice()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil)
		cached := NewCachedTargetDiscovery(discovery, WithRetryTargetsOnError(ctx, 5*time.Minute, time.Hour))

		cached.Refresh(ctx)
		cached.Refresh(ctx)
		time.Sleep(time.Hour)
		synctest.Wait()

		discovery.AssertNumberOfCalls(t, "DiscoverTargets", 3)
	})
}

func Test_retry_targets_on_error_stops_with_context(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		discovery := bareTargetDiscovery()
		discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
		cached := NewCachedTargetDiscovery(discovery, WithRetryTargetsOnError(ctx, 5*time.Minute, time.Hour))

		cached.Refresh(ctx)
		cancel()
		time.Sleep(time.Hour)
		synctest.Wait()

		discovery.AssertNumberOfCalls(t, "DiscoverTargets", 1)
	})
}
