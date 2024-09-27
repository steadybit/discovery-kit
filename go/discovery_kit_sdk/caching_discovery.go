// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/zmwangx/debounce"
	"runtime/debug"
	"sync"
	"time"
	"unique"
)

type discoverFn[T any] func(ctx context.Context) ([]T, error)

type CachedDiscovery[T any] struct {
	Discovery

	mu           sync.RWMutex
	lastModified time.Time
	supplier     discoverFn[T]
	data         []T
	err          error
}

type CachedDiscoveryOpt[T any] func(m *CachedDiscovery[T])

type CachedTargetDiscovery struct {
	CachedDiscovery[discovery_kit_api.Target]
}

type CachedDataEnrichmentDiscovery struct {
	CachedDiscovery[discovery_kit_api.EnrichmentData]
}

var (
	ErrDiscoveryTimeout = errors.New("discovery timed out")
)
var (
	_ TargetDiscovery         = (*CachedTargetDiscovery)(nil)
	_ Unwrapper               = (*CachedTargetDiscovery)(nil)
	_ EnrichmentDataDiscovery = (*CachedDataEnrichmentDiscovery)(nil)
	_ Unwrapper               = (*CachedDataEnrichmentDiscovery)(nil)
)

// NewCachedTargetDiscovery returns a caching target discovery.
func NewCachedTargetDiscovery(d TargetDiscovery, opts ...CachedDiscoveryOpt[discovery_kit_api.Target]) *CachedTargetDiscovery {
	c := &CachedTargetDiscovery{
		CachedDiscovery: CachedDiscovery[discovery_kit_api.Target]{
			Discovery: d,
			supplier:  recoverable(internStrings(internTargetStrings, d.DiscoverTargets)),
			data:      make([]discovery_kit_api.Target, 0),
		},
	}
	for _, opt := range opts {
		opt(&c.CachedDiscovery)
	}
	return c
}

func (c *CachedTargetDiscovery) DiscoverTargets(_ context.Context) ([]discovery_kit_api.Target, error) {
	return c.CachedDiscovery.Get()
}

// NewCachedEnrichmentDataDiscovery returns a caching enrichment data discovery.
func NewCachedEnrichmentDataDiscovery(d EnrichmentDataDiscovery, opts ...CachedDiscoveryOpt[discovery_kit_api.EnrichmentData]) *CachedDataEnrichmentDiscovery {
	c := &CachedDataEnrichmentDiscovery{
		CachedDiscovery: CachedDiscovery[discovery_kit_api.EnrichmentData]{
			Discovery: d,
			supplier:  recoverable(internStrings(internEnrichmentDataItemStrings, d.DiscoverEnrichmentData)),
			data:      make([]discovery_kit_api.EnrichmentData, 0),
		},
	}
	for _, opt := range opts {
		opt(&c.CachedDiscovery)
	}
	return c
}

func (c *CachedDataEnrichmentDiscovery) DiscoverEnrichmentData(_ context.Context) ([]discovery_kit_api.EnrichmentData, error) {
	return c.CachedDiscovery.Get()
}

func recoverable[T any](fn discoverFn[T]) discoverFn[T] {
	return func(ctx context.Context) (d []T, e error) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Msgf("discovery panic: %v\n %s", err, string(debug.Stack()))
			}
		}()
		return fn(ctx)
	}
}

func (c *CachedDiscovery[T]) Get() ([]T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data, c.err
}

func (c *CachedDiscovery[T]) LastModified() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastModified
}

func (c *CachedDiscovery[T]) Unwrap() interface{} {
	return c.Discovery
}

type UpdateFn[U any] func(U) (U, error)

func (c *CachedDiscovery[T]) Update(fn UpdateFn[[]T]) {
	lastModified := time.Now()
	log.Trace().Msg("updating discovery")
	data, err := fn(c.data)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastModified = lastModified
	c.data = data
	c.err = err
	if err == nil {
		log.Debug().TimeDiff("duration", time.Now(), c.lastModified).Int("count", len(data)).Msg("discovery updated")
	} else {
		var extensionError extension_kit.ExtensionError
		isExtensionError := errors.As(err, &extensionError)
		if isExtensionError {
			logEvent := log.Warn().TimeDiff("duration", time.Now(), c.lastModified).Str("title", extensionError.Title)
			if extensionError.Detail != nil {
				logEvent = logEvent.Str("detail", *extensionError.Detail)
			}
			if extensionError.Instance != nil {
				logEvent = logEvent.Str("instance", *extensionError.Instance)
			}
			logEvent.Msg("discovery update failed")
		} else {
			log.Warn().TimeDiff("duration", time.Now(), c.lastModified).Err(err).Msg("discovery update failed")
		}
	}
}

func (c *CachedDiscovery[T]) Refresh(ctx context.Context) {
	c.Update(func(_ []T) ([]T, error) {
		return c.supplier(ctx)
	})
}

// WithRefreshTargetsNow triggers a refresh of the cache immediately at creation time.
func WithRefreshTargetsNow() CachedDiscoveryOpt[discovery_kit_api.Target] {
	return WithRefreshNow[discovery_kit_api.Target]()
}

// WithRefreshEnrichmentDataNow triggers a refresh of the cache immediately at creation time.
func WithRefreshEnrichmentDataNow() CachedDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithRefreshNow[discovery_kit_api.EnrichmentData]()
}

// WithRefreshNow triggers a refresh of the cache immediately at creation time.
func WithRefreshNow[T any]() CachedDiscoveryOpt[T] {
	return func(m *CachedDiscovery[T]) {
		go func() {
			m.Refresh(context.Background())
		}()
	}
}

// WithRefreshTargetsTrigger triggers a refresh of the cache when an item on the channel is received and will stop when the context is canceled. The refreshes will be throttled by the given throttlePeriod.
func WithRefreshTargetsTrigger(ctx context.Context, ch <-chan struct{}, throttlePeriod time.Duration) CachedDiscoveryOpt[discovery_kit_api.Target] {
	return WithRefreshTrigger[discovery_kit_api.Target](ctx, ch, throttlePeriod)
}

// WithRefreshEnrichmentDataTrigger triggers a refresh of the cache when an item on the channel is received and will stop when the context is canceled. The refreshes will be throttled by the given throttlePeriod.
func WithRefreshEnrichmentDataTrigger(ctx context.Context, ch <-chan struct{}, throttlePeriod time.Duration) CachedDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithRefreshTrigger[discovery_kit_api.EnrichmentData](ctx, ch, throttlePeriod)
}

// WithRefreshTrigger triggers a refresh of the cache when an item on the channel is received and will stop when the context is canceled.
func WithRefreshTrigger[T any](ctx context.Context, ch <-chan struct{}, throttlePeriod time.Duration) CachedDiscoveryOpt[T] {
	return func(m *CachedDiscovery[T]) {
		fn := m.Refresh

		if throttlePeriod > 0 {
			debounced, _ := debounce.ThrottleWithCustomSignature(func(args ...interface{}) interface{} {
				m.Refresh(args[0].(context.Context))
				return nil
			}, throttlePeriod)
			fn = func(ctx context.Context) {
				debounced(ctx)
			}
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-ch:
					fn(ctx)
				}
			}
		}()
	}
}

// WithRefreshTargetsInterval triggers a refresh of the cache at the given interval and will stop when the context is canceled.
func WithRefreshTargetsInterval(ctx context.Context, interval time.Duration) CachedDiscoveryOpt[discovery_kit_api.Target] {
	return WithRefreshInterval[discovery_kit_api.Target](ctx, interval)
}

// WithRefreshEnrichmentDataInterval triggers a refresh of the cache at the given interval and will stop when the context is canceled.
func WithRefreshEnrichmentDataInterval(ctx context.Context, interval time.Duration) CachedDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithRefreshInterval[discovery_kit_api.EnrichmentData](ctx, interval)
}

// WithRefreshInterval triggers a refresh of the cache at the given interval and will stop when the context is canceled.
func WithRefreshInterval[T any](ctx context.Context, interval time.Duration) CachedDiscoveryOpt[T] {
	return func(m *CachedDiscovery[T]) {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(interval):
					m.Refresh(ctx)
				}
			}
		}()
	}
}

type UpdateFunc[D, U any] func(data D, update U) (D, error)

// WithTargetsUpdate triggers an updates the cache using the given function when an item on the channel is received and will stop when the context is canceled.
func WithTargetsUpdate[U any](ctx context.Context, ch <-chan U, fn UpdateFunc[[]discovery_kit_api.Target, U]) CachedDiscoveryOpt[discovery_kit_api.Target] {
	return WithUpdate[discovery_kit_api.Target, U](ctx, ch, fn)
}

// WithEnrichmentDataUpdate triggers an updates the cache using the given function when an item on the channel is received and will stop when the context is canceled.
func WithEnrichmentDataUpdate[U any](ctx context.Context, ch <-chan U, fn UpdateFunc[[]discovery_kit_api.EnrichmentData, U]) CachedDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithUpdate[discovery_kit_api.EnrichmentData, U](ctx, ch, fn)
}

func WithUpdate[T, U any](ctx context.Context, ch <-chan U, fn UpdateFunc[[]T, U]) CachedDiscoveryOpt[T] {
	return func(m *CachedDiscovery[T]) {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case update := <-ch:
					m.Update(func(data []T) ([]T, error) {
						return fn(data, update)
					})
				}
			}
		}()
	}
}

type result[T any] struct {
	data []T
	err  error
}

// WithTargetsRefreshTimeout decorates the supplier with a timeout.
func WithTargetsRefreshTimeout(timeout time.Duration) CachedDiscoveryOpt[discovery_kit_api.Target] {
	return WithRefreshTimeout[discovery_kit_api.Target](timeout)
}

// WithEnrichmentDataRefreshTimeout decorates the supplier with a timeout.
func WithEnrichmentDataRefreshTimeout(timeout time.Duration) CachedDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithRefreshTimeout[discovery_kit_api.EnrichmentData](timeout)
}

func WithRefreshTimeout[T any](timeout time.Duration) CachedDiscoveryOpt[T] {
	return func(m *CachedDiscovery[T]) {
		supplier := m.supplier

		m.supplier = func(ctx context.Context) ([]T, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			ch := make(chan result[T], 1)
			go func() {
				data, err := supplier(ctx)
				ch <- result[T]{data: data, err: err}
			}()

			select {
			case <-ctx.Done():
				return nil, ErrDiscoveryTimeout
			case r := <-ch:
				return r.data, r.err
			}
		}
	}
}

type makeHandleFunc func(s string) string

func internStrings[T any](fnInternItem func(makeHandleFunc, *T) error, fn discoverFn[T]) discoverFn[T] {
	return func(ctx context.Context) ([]T, error) {
		data, err := fn(ctx)
		if err != nil || len(data) == 0 {
			return data, err
		}

		//keep handles until all are process so gc won't remove them and break the deduplication
		handles := make([]unique.Handle[string], len(data))
		makeHandle := func(s string) string {
			h := unique.Make(s)
			handles = append(handles, h)
			return h.Value()
		}

		for i := range data {
			err = errors.Join(fnInternItem(makeHandle, &data[i]))
		}
		return data, err
	}
}

func internEnrichmentDataItemStrings(makeHandle makeHandleFunc, data *discovery_kit_api.EnrichmentData) error {
	if data == nil {
		return nil
	}

	data.Id = makeHandle(data.Id)
	data.EnrichmentDataType = makeHandle(data.EnrichmentDataType)
	for key, values := range data.Attributes {
		for i, value := range values {
			values[i] = makeHandle(value)
		}
		data.Attributes[makeHandle(key)] = values
	}

	return nil
}

func internTargetStrings(makeHandle makeHandleFunc, data *discovery_kit_api.Target) error {
	if data == nil {
		return nil
	}

	data.Id = makeHandle(data.Id)
	data.TargetType = makeHandle(data.TargetType)
	for key, values := range data.Attributes {
		for i, value := range values {
			values[i] = makeHandle(value)
		}
		data.Attributes[makeHandle(key)] = values
	}

	return nil
}
