// Copyright 2025 steadybit GmbH. All rights reserved.

package discovery_kit_sdk

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
	"unique"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/zmwangx/debounce"
)

type discoverFn[T any] func(ctx context.Context) ([]T, error)

type CachedDiscovery[T any] struct {
	Discovery

	mu           sync.RWMutex
	lastModified time.Time
	supplier     discoverFn[T]
	data         []T
	err          error
	lastSuccess  time.Time
	keepDataFor  time.Duration
	retry        *retryOnError
	start        []func()
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
	c.apply(opts)
	return c
}

// apply starts the background refreshes only after all options are applied, so the first refresh
// already uses every option regardless of their order.
func (c *CachedDiscovery[T]) apply(opts []CachedDiscoveryOpt[T]) {
	for _, opt := range opts {
		opt(c)
	}
	for _, start := range c.start {
		start()
	}
	c.start = nil
}

func (c *CachedDiscovery[T]) onStart(fn func()) {
	c.start = append(c.start, fn)
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
	c.apply(opts)
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
				// Surface the panic as an error so the cache records a failed discovery
				// instead of a successful empty result (which would publish an empty
				// target/enrichment set and advance the ETag as if it were legitimate).
				e = fmt.Errorf("discovery panicked: %v", err)
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

func (c *CachedDiscovery[T]) Unwrap() any {
	return c.Discovery
}

type UpdateFn[U any] func(U) (U, error)

func (c *CachedDiscovery[T]) Update(fn UpdateFn[[]T]) {
	c.update(fn)
}

func (c *CachedDiscovery[T]) update(fn UpdateFn[[]T]) error {
	started := time.Now()
	log.Trace().Msg("updating discovery")

	c.mu.RLock()
	current := c.data
	c.mu.RUnlock()

	data, err := fn(current)

	c.mu.Lock()
	defer c.mu.Unlock()
	if err == nil {
		c.lastSuccess = started
	} else if c.keepDataFor > 0 && !c.lastSuccess.IsZero() && time.Since(c.lastSuccess) <= c.keepDataFor {
		// lastModified stays untouched, so the ETag tells the agent nothing changed.
		updateFailedEvent(err, started).Time("kept_from", c.lastSuccess).Msg("discovery update failed, keeping the last successful result")
		return err
	}
	c.lastModified = started
	c.data = data
	c.err = err
	if err == nil {
		log.Debug().TimeDiff("duration", time.Now(), started).Int("count", len(data)).Msg("discovery updated")
	} else {
		updateFailedEvent(err, started).Msg("discovery update failed")
	}
	return err
}

func updateFailedEvent(err error, started time.Time) *zerolog.Event {
	event := log.Warn().TimeDiff("duration", time.Now(), started)
	var extensionError extension_kit.ExtensionError
	if !errors.As(err, &extensionError) {
		return event.Err(err)
	}
	event = event.Str("title", extensionError.Title)
	if extensionError.Detail != nil {
		event = event.Str("detail", *extensionError.Detail)
	}
	if extensionError.Instance != nil {
		event = event.Str("instance", *extensionError.Instance)
	}
	return event
}

func (c *CachedDiscovery[T]) Refresh(ctx context.Context) {
	err := c.update(func(_ []T) ([]T, error) {
		return c.supplier(ctx)
	})
	if c.retry != nil {
		c.retry.refreshed(err)
	}
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
		m.onStart(func() {
			go m.Refresh(context.Background())
		})
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
			debounced, _ := debounce.ThrottleWithCustomSignature(func(args ...any) any {
				m.Refresh(args[0].(context.Context))
				return nil
			}, throttlePeriod)
			fn = func(ctx context.Context) {
				debounced(ctx)
			}
		}

		m.onStart(func() {
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
		})
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
		m.onStart(func() {
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
		})
	}
}

// WithKeepTargetsOnError keeps serving the targets of the last successful refresh when a refresh fails, for at most maxAge since that refresh.
func WithKeepTargetsOnError(maxAge time.Duration) CachedDiscoveryOpt[discovery_kit_api.Target] {
	return WithKeepDataOnError[discovery_kit_api.Target](maxAge)
}

// WithKeepEnrichmentDataOnError keeps serving the enrichment data of the last successful refresh when a refresh fails, for at most maxAge since that refresh.
func WithKeepEnrichmentDataOnError(maxAge time.Duration) CachedDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithKeepDataOnError[discovery_kit_api.EnrichmentData](maxAge)
}

// WithKeepDataOnError keeps serving the data of the last successful refresh when a refresh fails, for at most maxAge since that refresh.
// Without it, a failed refresh replaces the data with the error until the next refresh succeeds, which is costly for discoveries
// with a long refresh interval against an occasionally failing API. Once maxAge is exceeded, or if no refresh has succeeded yet,
// the error is served, so a permanent failure is not hidden behind outdated data.
func WithKeepDataOnError[T any](maxAge time.Duration) CachedDiscoveryOpt[T] {
	return func(m *CachedDiscovery[T]) {
		m.keepDataFor = maxAge
	}
}

// WithRetryTargetsOnError retries a failed refresh with exponential backoff, see WithRetryOnError.
func WithRetryTargetsOnError(ctx context.Context, minDelay, maxDelay time.Duration) CachedDiscoveryOpt[discovery_kit_api.Target] {
	return WithRetryOnError[discovery_kit_api.Target](ctx, minDelay, maxDelay)
}

// WithRetryEnrichmentDataOnError retries a failed refresh with exponential backoff, see WithRetryOnError.
func WithRetryEnrichmentDataOnError(ctx context.Context, minDelay, maxDelay time.Duration) CachedDiscoveryOpt[discovery_kit_api.EnrichmentData] {
	return WithRetryOnError[discovery_kit_api.EnrichmentData](ctx, minDelay, maxDelay)
}

// WithRetryOnError retries a failed refresh after minDelay, doubling the delay on every further failure, instead of waiting for
// the next regular refresh. Retrying stops once the delay would exceed maxDelay, so pass the refresh interval as maxDelay to have
// the regular refreshes take over from there without retries running in between. A successful refresh cancels a pending retry
// and resets the delay. Retries stop when the context is canceled.
func WithRetryOnError[T any](ctx context.Context, minDelay, maxDelay time.Duration) CachedDiscoveryOpt[T] {
	return func(m *CachedDiscovery[T]) {
		m.retry = &retryOnError{
			minDelay: minDelay,
			maxDelay: maxDelay,
			refresh: func() {
				if ctx.Err() == nil {
					m.Refresh(ctx)
				}
			},
		}
	}
}

type retryOnError struct {
	minDelay time.Duration
	maxDelay time.Duration
	refresh  func()

	mu    sync.Mutex
	delay time.Duration
	timer *time.Timer
}

func (r *retryOnError) refreshed(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err == nil {
		if r.timer != nil {
			r.timer.Stop()
			r.timer = nil
		}
		r.delay = 0
		return
	}
	// A refresh from another source (e.g. the interval) failing while a retry is pending must not schedule a second one.
	if r.timer != nil || r.delay > r.maxDelay {
		return
	}
	r.delay = max(2*r.delay, r.minDelay)
	if r.delay > r.maxDelay {
		return
	}
	r.timer = time.AfterFunc(r.delay, func() {
		r.mu.Lock()
		r.timer = nil
		r.mu.Unlock()
		r.refresh()
	})
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
		m.onStart(func() {
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
		})
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

func internStrings[T any](fnInternItem func(makeHandleFunc, T) T, fn discoverFn[T]) discoverFn[T] {
	return func(ctx context.Context) ([]T, error) {
		data, err := fn(ctx)
		if err != nil || len(data) == 0 {
			return data, err
		}

		//keep handles until all are processed so gc won't remove them and break the deduplication
		handles := make([]unique.Handle[string], len(data))
		makeHandle := func(s string) string {
			h := unique.Make(s)
			handles = append(handles, h)
			return h.Value()
		}

		newData := make([]T, len(data))
		for i := range data {
			newData[i] = fnInternItem(makeHandle, data[i])
		}
		return newData, err
	}
}

func internEnrichmentDataItemStrings(makeHandle makeHandleFunc, data discovery_kit_api.EnrichmentData) discovery_kit_api.EnrichmentData {
	return discovery_kit_api.EnrichmentData{
		Id:                 makeHandle(data.Id),
		EnrichmentDataType: makeHandle(data.EnrichmentDataType),
		Attributes:         internAttributes(makeHandle, data.Attributes),
	}
}

func internTargetStrings(makeHandle makeHandleFunc, data discovery_kit_api.Target) discovery_kit_api.Target {
	return discovery_kit_api.Target{
		Id:         makeHandle(data.Id),
		Label:      makeHandle(data.Label),
		TargetType: makeHandle(data.TargetType),
		Attributes: internAttributes(makeHandle, data.Attributes),
	}
}

func internAttributes(makeHandle makeHandleFunc, attributes map[string][]string) map[string][]string {
	var newAttributes = make(map[string][]string, len(attributes))
	for key, values := range attributes {
		var newValues = make([]string, len(values))
		for i, value := range values {
			newValues[i] = makeHandle(value)
		}
		newAttributes[makeHandle(key)] = newValues
	}
	return newAttributes
}
