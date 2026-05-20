# Changelog

## 1.3.4

- fix: fatal error introduced with 1.3.3: concurrent map iteration and map write when STEADYBIT_EXTENSION_DISCOVERY_GROUP is set. The group attribute is now applied once per refresh in the supplier chain of the cache (CachedTargetDiscovery / CachedDataEnrichmentDiscovery), where the attribute map is freshly allocated and not yet shared with any reader. Previously it was written into the cache's map at every HTTP request, which caused the panic under concurrent requests.
- feat: expose the group-attribute injection as a public `CachedDiscoveryOpt`: `WithGroupAttribute[T]` plus typed wrappers `WithTargetsGroupAttribute` / `WithEnrichmentDataGroupAttribute`. Auto-applied by `NewCachedTargetDiscovery` / `NewCachedEnrichmentDataDiscovery`, so existing callers need no changes.

## 1.3.3

- Support setting "STEADYBIT_EXTENSION_DISCOVERY_GROUP" environment variable to set the "steadybit.group" attribute for all discovered targets and enrichment data records.

## 1.3.2

- Update dependencies

## 1.3.1

- fix: fatal error: concurrent map iteration and map write

## 1.3.0

- Update dependencies (golang 1.24)

## 1.2.2, 1.2.3

- Added a check for duplicate targets in the discovery data

## 1.2.1

- Fix: add missing Target.Label to string interning

## 1.2.0

- Update to go 1.23
- Intern the discovery data strings by default

## 1.1.1

- add http request to context

## 1.1.0

- Update to discovery_kit_api 1.6.0

## 1.0.7

- additional logging for extension errors during discovery updates

## 1.0.6

- code cleanup

## 1.0.5

- fix: caching discovery usage of write lock
- update dependencies

## 1.0.4

- update to discovery_kit_test 1.1.2

## 1.0.3

- add WithRefreshTimeout option for cached discovery

## 1.0.2

- fix target enrichment rule http adapter

## 1.0.1

- add debug logging when refreshing discovery data

## 1.0.0

- Initial release

