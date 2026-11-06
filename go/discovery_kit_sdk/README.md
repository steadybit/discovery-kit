# DsicoveryKit Go SDK

This module contains helpers and interfaces which will help you to implement discoveries using
the [discovery_kit go api](https://github.com/steadybit/discovery-kit/tree/main/go/discovery_kit_api).

The module encapsulates the following technical aspects and provides helpers for the following elements:

- The SDK will wrap around your `describe` call and provide meaningful defaults for your endpoint definitions.
- Caching and async discovery execution to decouple the discovery execution from the HTTP requests. 

## Installation

Add the following to your `go.mod` file:

```
go get github.com/steadybit/discovery-kit/go/discovery_kit_sdk
```

## Usage

1. Implement at least the `discovery_kit_sdk.TargetDiscovery` or `discovery_kit_sdk.EnrichmentDataDisocvery` interface

2. Implement other interfaces if you need them:
    - `discovery_kit_sdk.TargetDescriber`
    - `discovery_kit_sdk.AttributesDescriber`
    - `discovery_kit_sdk.EnrichmentRuleContributor`

3. Register your discovery:
   ```go
   discovery_kit_sdk.Register(NewMyCustomDiscovery())
   ```

4. Add your registered discoveries to the index endpoint of your extension:
   ```go
   exthttp.RegisterHttpHandler("/discoveries", exthttp.GetterAsHandler(discovery_kit_sdk.GetDiscoveryList))
   ```

## Caching / Async Discovery

If you implement the `TargetDiscovery` / `EnrichmentDataDiscovery` in a straightforward fashion, then the discovery is executed synchronously for each HTTP request.
You can decouple this by decorating your discovery using the `NewCachedTargetDiscovery` / `NewCachedEnrichmentDataDiscovery` functions. You pass your discovery and options to control the refreshing of the data.

```go
	discovery := &jvmDiscovery{}
	return discovery_kit_sdk.NewCachedTargetDiscovery(discovery,
		discovery_kit_sdk.WithRefreshTargetsNow(),
		discovery_kit_sdk.WithRefreshTargetsInterval(context.Background(), 30*time.Second),
	)
```

You have various options to refresh periodically once on the trigger. Also, it will help recover from any panic upon discovery.

By default, a failed refresh replaces the cached data with the error until a refresh succeeds again. For discoveries with a long
refresh interval against an external API, a single failed call would then hide all targets for the whole interval. Two opt-in
options soften this:

- `WithKeepTargetsOnError(maxAge)` keeps serving the last successful result when a refresh fails, as long as it is at most
  `maxAge` old. Without any successful refresh, or past `maxAge`, the error is served.
- `WithRetryTargetsOnError(ctx, minDelay, maxDelay)` retries a failed refresh after `minDelay`, doubling the delay on every
  further failure until it would exceed `maxDelay`. Pass the refresh interval as `maxDelay`, so the regular refreshes take over
  from there.

```go
	ctx := context.Background()
	return discovery_kit_sdk.NewCachedTargetDiscovery(discovery,
		discovery_kit_sdk.WithRefreshTargetsNow(),
		discovery_kit_sdk.WithRefreshTargetsInterval(ctx, 3*time.Hour),
		discovery_kit_sdk.WithKeepTargetsOnError(24*time.Hour),
		discovery_kit_sdk.WithRetryTargetsOnError(ctx, 5*time.Minute, 3*time.Hour),
	)
```

Both have `EnrichmentData` and generic variants.

## Filtering targets out of discovery

Operators can exclude targets with a query, configured by environment variable. Nothing is
required of the discovery implementation — see [Target Filtering](/docs/target-filtering.md).
