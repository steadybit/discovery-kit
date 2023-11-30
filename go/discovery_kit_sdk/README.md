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



