// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"runtime/debug"
	"sync"
	"time"
)

type CachingDiscovery[T any] struct {
	Discovery

	mu           sync.RWMutex
	lastModified time.Time
	supplier     func(ctx context.Context) []T
	data         []T
}

type CachingDiscoveryOpt[T any] func(m *CachingDiscovery[T])

type CachingTargetDiscovery struct {
	CachingDiscovery[discovery_kit_api.Target]
}

type CachingDataEnrichmentDiscovery struct {
	CachingDiscovery[discovery_kit_api.EnrichmentData]
}

// CachedTargetDiscovery returns a caching target discovery.
func CachedTargetDiscovery(d TargetDiscovery, opts ...CachingDiscoveryOpt[discovery_kit_api.Target]) *CachingTargetDiscovery {
	c := &CachingTargetDiscovery{
		CachingDiscovery: CachingDiscovery[discovery_kit_api.Target]{
			Discovery: d,
			supplier:  recoverable(d.DiscoverTargets),
			data:      make([]discovery_kit_api.Target, 0),
		},
	}
	for _, opt := range opts {
		opt(&c.CachingDiscovery)
	}
	return c
}

// CachedEnrichmentDataDiscovery returns a caching enrichment data discovery.
func CachedEnrichmentDataDiscovery(d EnrichmentDataDiscovery, opts ...CachingDiscoveryOpt[discovery_kit_api.EnrichmentData]) *CachingDataEnrichmentDiscovery {
	c := &CachingDataEnrichmentDiscovery{
		CachingDiscovery: CachingDiscovery[discovery_kit_api.EnrichmentData]{
			Discovery: d,
			supplier:  recoverable(d.DiscoverEnrichmentData),
			data:      make([]discovery_kit_api.EnrichmentData, 0),
		},
	}
	for _, opt := range opts {
		opt(&c.CachingDiscovery)
	}
	return c
}

func recoverable[T any](fn func(ctx context.Context) T) func(ctx context.Context) T {
	return func(ctx context.Context) T {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msgf("discovery panic: %v\n %s", err, string(debug.Stack()))
			}
		}()
		return fn(ctx)
	}
}

func (c *CachingTargetDiscovery) DiscoverTargets(_ context.Context) []discovery_kit_api.Target {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data
}

func (c *CachingDataEnrichmentDiscovery) DiscoverEnrichmentData(_ context.Context) []discovery_kit_api.EnrichmentData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data
}

func (c *CachingDiscovery[T]) LastModified() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastModified
}

func (c *CachingDiscovery[T]) Unwrap() interface{} {
	return c.Discovery
}

type mapper[U any] func(U) U

func (c *CachingDiscovery[T]) update(fn mapper[[]T]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastModified = time.Now()
	c.data = fn(c.data)
}

func (c *CachingDiscovery[T]) refresh(ctx context.Context) {
	c.update(func(_ []T) []T {
		return c.supplier(ctx)
	})
}

// WithRefreshTargetsNow triggers a refresh of the cache immediately at creation time.
func WithRefreshTargetsNow() CachingDiscoveryOpt[discovery_kit_api.Target] {
	return WithRefreshNow[discovery_kit_api.Target]()
}

// WithRefreshEnrichmentDataNow triggers a refresh of the cache immediately at creation time.
func WithRefreshEnrichmentDataNow() CachingDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithRefreshNow[discovery_kit_api.EnrichmentData]()
}

// WithRefreshNow triggers a refresh of the cache immediately at creation time.
func WithRefreshNow[T any]() CachingDiscoveryOpt[T] {
	return func(m *CachingDiscovery[T]) {
		go func() {
			m.refresh(context.Background())
		}()
	}
}

// WithRefreshTargetsTrigger triggers a refresh of the cache when an item on the channel is received and will stop when the context is canceled.
func WithRefreshTargetsTrigger(ctx context.Context, ch <-chan struct{}) CachingDiscoveryOpt[discovery_kit_api.Target] {
	return WithRefreshTrigger[discovery_kit_api.Target](ctx, ch)
}

// WithRefreshEnrichmentDataTrigger triggers a refresh of the cache when an item on the channel is received and will stop when the context is canceled.
func WithRefreshEnrichmentDataTrigger(ctx context.Context, ch <-chan struct{}) CachingDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithRefreshTrigger[discovery_kit_api.EnrichmentData](ctx, ch)
}

// WithRefreshTrigger triggers a refresh of the cache when an item on the channel is received and will stop when the context is canceled.
func WithRefreshTrigger[T any](ctx context.Context, ch <-chan struct{}) CachingDiscoveryOpt[T] {
	return func(m *CachingDiscovery[T]) {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-ch:
					m.refresh(ctx)
				}
			}
		}()
	}
}

// WithRefreshTargetsInterval triggers a refresh of the cache at the given interval and will stop when the context is canceled.
func WithRefreshTargetsInterval(ctx context.Context, interval time.Duration) CachingDiscoveryOpt[discovery_kit_api.Target] {
	return WithRefreshInterval[discovery_kit_api.Target](ctx, interval)
}

// WithRefreshEnrichmentDataInterval triggers a refresh of the cache at the given interval and will stop when the context is canceled.
func WithRefreshEnrichmentDataInterval(ctx context.Context, interval time.Duration) CachingDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithRefreshInterval[discovery_kit_api.EnrichmentData](ctx, interval)
}

// WithRefreshInterval triggers a refresh of the cache at the given interval and will stop when the context is canceled.
func WithRefreshInterval[T any](ctx context.Context, interval time.Duration) CachingDiscoveryOpt[T] {
	return func(m *CachingDiscovery[T]) {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(interval):
					m.refresh(ctx)
				}
			}
		}()
	}
}

type UpdateFunc[D, U any] func(data D, update U) D

// WithTargetsUpdate triggers an updates the cache using the given function when an item on the channel is received and will stop when the context is canceled.
func WithTargetsUpdate[U any](ctx context.Context, ch <-chan U, fn UpdateFunc[[]discovery_kit_api.Target, U]) CachingDiscoveryOpt[discovery_kit_api.Target] {
	return WithUpdate[discovery_kit_api.Target, U](ctx, ch, fn)
}

// WithEnrichmentDataUpdate triggers an updates the cache using the given function when an item on the channel is received and will stop when the context is canceled.
func WithEnrichmentDataUpdate[U any](ctx context.Context, ch <-chan U, fn UpdateFunc[[]discovery_kit_api.EnrichmentData, U]) CachingDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithUpdate[discovery_kit_api.EnrichmentData, U](ctx, ch, fn)
}

func WithUpdate[T, U any](ctx context.Context, ch <-chan U, fn UpdateFunc[[]T, U]) CachingDiscoveryOpt[T] {
	return func(m *CachingDiscovery[T]) {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case update := <-ch:
					m.update(func(data []T) []T {
						return fn(data, update)
					})
				}
			}
		}()
	}
}
