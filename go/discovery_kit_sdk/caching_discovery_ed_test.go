// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"errors"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
	"unsafe"
)

func Test_enrichmentData_caching(t *testing.T) {
	ctx := context.Background()

	discovery := newMockEnrichmentDataDiscovery()
	cached := NewCachedEnrichmentDataDiscovery(discovery)

	cached.Refresh(context.Background())
	first, _ := cached.DiscoverEnrichmentData(ctx)
	second, _ := cached.DiscoverEnrichmentData(ctx)

	assert.Equal(t, first, second)

	discovery.AssertNumberOfCalls(t, "DiscoverEnrichmentData", 1)
}

func Test_enrichmentData_timeout(t *testing.T) {
	ctx := context.Background()

	discovery := newMockEnrichmentDataDiscovery()
	cached := NewCachedEnrichmentDataDiscovery(discovery, WithEnrichmentDataRefreshTimeout(1*time.Second))

	discovery.On("DiscoverEnrichmentData", mock.Anything).Unset()
	discovery.On("DiscoverEnrichmentData", mock.Anything).Return([]discovery_kit_api.EnrichmentData{{}}, nil).Once()
	call := discovery.On("DiscoverEnrichmentData", mock.Anything).Return([]discovery_kit_api.EnrichmentData{}, nil).Once()
	call.RunFn = func(args mock.Arguments) {
		time.Sleep(5 * time.Second)
	}
	discovery.On("DiscoverEnrichmentData", mock.Anything).Return([]discovery_kit_api.EnrichmentData{{}}, nil).Once()

	cached.Refresh(ctx)
	first, _ := cached.DiscoverEnrichmentData(ctx)
	assert.Len(t, first, 1)

	cached.Refresh(ctx)
	_, secondErr := cached.DiscoverEnrichmentData(ctx)
	assert.ErrorIs(t, secondErr, ErrDiscoveryTimeout)

	cached.Refresh(ctx)
	third, _ := cached.DiscoverEnrichmentData(ctx)
	assert.Len(t, third, 1)
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

	trigger := func() {
		ch <- struct{}{}
	}

	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)
	data, err := cached.DiscoverEnrichmentData(ctx)
	assert.NoError(t, err)
	assert.Len(t, data, 1)

	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)
	data, err = cached.DiscoverEnrichmentData(ctx)
	assert.Error(t, err)
	assert.Len(t, data, 0)

	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)
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
	waitForLastModifiedChanges(t, &cached.CachedDiscovery)
	first, _ := cached.DiscoverEnrichmentData(ctx)
	second, _ := cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first, _ = cached.DiscoverEnrichmentData(ctx)
	waitForLastModifiedChanges(t, &cached.CachedDiscovery)
	second, _ = cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)
	waitForLastModifiedChanges(t, &cached.CachedDiscovery)
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
	trigger := func() {
		ch <- struct{}{}
	}
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)

	first, _ := cached.DiscoverEnrichmentData(ctx)
	second, _ := cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)

	//should refresh cache
	first, _ = cached.DiscoverEnrichmentData(ctx)
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, trigger)
	second, _ = cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)

	//should not refresh cache after cancel
	cancel()
	first, _ = cached.DiscoverEnrichmentData(ctx)
	time.Sleep(200 * time.Millisecond)
	second, _ = cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)
}

func Test_enrichmentData_cache_trigger_throttle(t *testing.T) {
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
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, func() {
		ch <- struct{}{}
	})
	second, _ := cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)
	discovery.AssertNumberOfCalls(t, "DiscoverEnrichmentData", 2)
}

func Test_enrichmentData_cache_update(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	discovery := newMockEnrichmentDataDiscovery()
	ch := make(chan string)
	updateFn := func(data []discovery_kit_api.EnrichmentData, update string) ([]discovery_kit_api.EnrichmentData, error) {
		if update == "clear" {
			return []discovery_kit_api.EnrichmentData{}, nil
		}
		return data, nil
	}
	cached := NewCachedEnrichmentDataDiscovery(discovery,
		WithEnrichmentDataUpdate(ctx, ch, updateFn),
	)

	//should cache
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, func() {
		cached.Refresh(context.Background())
	})
	first, _ := cached.DiscoverEnrichmentData(ctx)
	second, _ := cached.DiscoverEnrichmentData(ctx)
	assert.Equal(t, first, second)

	//should update cache
	first, _ = cached.DiscoverEnrichmentData(ctx)
	triggerAndWaitForUpdate(t, &cached.CachedDiscovery, func() {
		ch <- "clear"
	})
	second, _ = cached.DiscoverEnrichmentData(ctx)
	assert.NotEqual(t, first, second)
	assert.Empty(t, second)
}

func Test_enrichment_data_string_interning(t *testing.T) {
	largeString := "ID: this is a very large string which should get unshared"
	ctx := context.Background()

	discovery := newMockEnrichmentDataDiscovery()
	cached := NewCachedEnrichmentDataDiscovery(discovery)

	discovery.On("DiscoverEnrichmentData", mock.Anything).Unset()
	discovery.On("DiscoverEnrichmentData", ctx).Return([]discovery_kit_api.EnrichmentData{{
		Id: largeString[:2],
		Attributes: map[string][]string{
			largeString[:2]: {largeString[4:]},
		},
	}}, nil)
	cached.Refresh(ctx)
	data, _ := cached.DiscoverEnrichmentData(ctx)

	assert.Equal(t, "ID", data[0].Id)
	assert.Equal(t, []string{"this is a very large string which should get unshared"}, data[0].Attributes["ID"])

	assertSliceNotShared(t, largeString, data[0].Id)
	for _, datum := range data {
		for key, values := range datum.Attributes {
			assertSliceNotShared(t, largeString, key)
			for _, value := range values {
				assertSliceNotShared(t, largeString, value)
			}
		}
	}
}

func waitForLastModifiedChanges(t *testing.T, p p) {
	triggerAndWaitForUpdate(t, p, nil)
}

func triggerAndWaitForUpdate(t *testing.T, cached p, trigger func()) {
	t.Helper()

	lm := cached.LastModified()
	if trigger != nil {
		trigger()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for {
		if lm != cached.LastModified() {
			break
		}
		if ctx.Err() != nil {
			t.Fatalf("timeout waiting for last modified changes")
		}
		time.Sleep(10 * time.Millisecond)
	}

}

func assertSliceNotShared(t *testing.T, haystack string, needle string) {
	t.Helper()

	pHayBegin := uintptr(unsafe.Pointer(unsafe.StringData(haystack)))
	pHayEnd := uintptr(unsafe.Add(unsafe.Pointer(unsafe.StringData(haystack)), len(haystack)))
	pNeedle := uintptr(unsafe.Pointer(unsafe.StringData(needle)))

	if pNeedle >= pHayBegin && pNeedle < pHayEnd {
		t.Errorf("slice is shared")
	}
}
