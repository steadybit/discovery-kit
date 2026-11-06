// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errDiscovery = errors.New("api returned 500")

func failingTargetDiscovery() *MockTargetDiscovery {
	discovery := newMockTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Unset()
	return discovery
}

// countingDiscovery counts calls race-free, unlike reading the mock's Calls while retries run.
type countingDiscovery struct {
	*MockTargetDiscovery
	calls atomic.Int32
}

func (d *countingDiscovery) DiscoverTargets(ctx context.Context) ([]discovery_kit_api.Target, error) {
	d.calls.Add(1)
	return d.MockTargetDiscovery.DiscoverTargets(ctx)
}

func Test_keep_targets_on_error(t *testing.T) {
	ctx := t.Context()
	discovery := failingTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil).Once()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
	cached := NewCachedTargetDiscovery(discovery, WithKeepTargetsOnError(time.Hour))

	cached.Refresh(ctx)
	lastModified := cached.LastModified()
	cached.Refresh(ctx)

	targets, err := cached.DiscoverTargets(ctx)
	require.NoError(t, err)
	assert.Equal(t, []discovery_kit_api.Target{{Id: "a", Attributes: map[string][]string{}}}, targets)
	assert.Equal(t, lastModified, cached.LastModified(), "keeping the data must not change the ETag")
}

func Test_keep_targets_on_error_reports_error_without_previous_success(t *testing.T) {
	ctx := t.Context()
	discovery := failingTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
	cached := NewCachedTargetDiscovery(discovery, WithKeepTargetsOnError(time.Hour))

	cached.Refresh(ctx)

	_, err := cached.DiscoverTargets(ctx)
	assert.ErrorIs(t, err, errDiscovery)
}

func Test_keep_targets_on_error_reports_error_once_too_old(t *testing.T) {
	ctx := t.Context()
	discovery := failingTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil).Once()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
	cached := NewCachedTargetDiscovery(discovery, WithKeepTargetsOnError(time.Hour))

	cached.Refresh(ctx)
	cached.lastSuccess = time.Now().Add(-2 * time.Hour)
	cached.Refresh(ctx)

	_, err := cached.DiscoverTargets(ctx)
	assert.ErrorIs(t, err, errDiscovery)
}

// A failed refresh used to leave the error in place until the next interval refresh, hours away for
// discoveries against rate-limited SaaS APIs.
func Test_retry_targets_on_error_recovers_before_next_interval(t *testing.T) {
	ctx := t.Context()
	discovery := failingTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery).Once()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil)
	counted := &countingDiscovery{MockTargetDiscovery: discovery}
	cached := NewCachedTargetDiscovery(counted,
		WithRefreshTargetsNow(),
		WithRefreshTargetsInterval(ctx, time.Hour),
		WithRetryTargetsOnError(ctx, 10*time.Millisecond, time.Hour),
	)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		targets, err := cached.DiscoverTargets(ctx)
		assert.NoError(c, err)
		assert.Len(c, targets, 1)
	}, time.Second, 5*time.Millisecond)
	assert.Equal(t, int32(2), counted.calls.Load())
}

func Test_retry_targets_on_error_backs_off_until_max_delay(t *testing.T) {
	ctx := t.Context()
	discovery := failingTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
	counted := &countingDiscovery{MockTargetDiscovery: discovery}
	cached := NewCachedTargetDiscovery(counted, WithRetryTargetsOnError(ctx, 10*time.Millisecond, 30*time.Millisecond))

	cached.Refresh(ctx)

	// Retries after 10ms and 20ms; the next delay of 40ms exceeds the 30ms maximum.
	assert.Eventually(t, func() bool { return counted.calls.Load() == 3 }, time.Second, 5*time.Millisecond)
	time.Sleep(80 * time.Millisecond)
	assert.Equal(t, int32(3), counted.calls.Load())
}

func Test_retry_targets_on_error_is_canceled_by_successful_refresh(t *testing.T) {
	ctx := t.Context()
	discovery := failingTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery).Once()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil)
	counted := &countingDiscovery{MockTargetDiscovery: discovery}
	cached := NewCachedTargetDiscovery(counted, WithRetryTargetsOnError(ctx, 30*time.Millisecond, time.Hour))

	cached.Refresh(ctx)
	cached.Refresh(ctx)

	time.Sleep(60 * time.Millisecond)
	assert.Equal(t, int32(2), counted.calls.Load())
}

func Test_retry_targets_on_error_does_not_stack_retries(t *testing.T) {
	ctx := t.Context()
	discovery := failingTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery).Twice()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target{{Id: "a"}}, nil)
	counted := &countingDiscovery{MockTargetDiscovery: discovery}
	cached := NewCachedTargetDiscovery(counted, WithRetryTargetsOnError(ctx, 20*time.Millisecond, time.Hour))

	cached.Refresh(ctx)
	cached.Refresh(ctx)

	assert.Eventually(t, func() bool { return counted.calls.Load() == 3 }, time.Second, 5*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(3), counted.calls.Load())
}

func Test_retry_targets_on_error_stops_with_context(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	discovery := failingTargetDiscovery()
	discovery.On("DiscoverTargets", mock.Anything).Return([]discovery_kit_api.Target(nil), errDiscovery)
	counted := &countingDiscovery{MockTargetDiscovery: discovery}
	cached := NewCachedTargetDiscovery(counted, WithRetryTargetsOnError(ctx, 10*time.Millisecond, time.Hour))

	cached.Refresh(ctx)
	cancel()

	time.Sleep(40 * time.Millisecond)
	assert.Equal(t, int32(1), counted.calls.Load())
}
