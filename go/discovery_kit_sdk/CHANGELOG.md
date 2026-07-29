# Changelog

## 1.4.1

- Requires `extension-kit` v1.11.1 (up from v1.11.0), for the `extquery` package the filter below is built on.
- feat: operators can exclude targets from discovery inside the extension, using the same query language as environments and blast radii. `STEADYBIT_EXTENSION_DISCOVERY_EXCLUDE_QUERY` drops matching targets and `STEADYBIT_EXTENSION_DISCOVERY_INCLUDE_QUERY` keeps only matching ones; both accept a `_FILE` suffix to read the query from a file. A query applies to every discovery of the extension — narrow it to one with `target.type="..."` in the query itself. Filtered targets never leave the extension, so they cost nothing to transfer, ingest or store. A malformed query stops the extension at start up rather than being ignored. See the README.

## 1.4.0

- **Requires Go 1.26.5** (up from 1.25.0) and `extension-kit` v1.11.0 (up from v1.10.4). Extensions must raise their own `go` directive and toolchain to build against this version.
- feat: `Register` and `ClearRegisteredDiscoveries` now call `exthttp.BumpRevision()`, so the extension index ETag reflects the set of registered discoveries, target describers, attribute describers and enrichment rules. The agent's index-response cache invalidates on registration changes without needing a process restart. The Discover endpoint's own ETag is unchanged.
- fix: data race in `CachedDiscovery.Update` — the previous target/enrichment slice was read while building the new one without holding the lock, racing a concurrent `Update` writing it. It surfaced as a `-race` test failure when a discovery registered with both `WithRefreshTargetsNow` and a short `WithRefreshTargetsInterval` ran its two initial refreshes concurrently. The prior slice is now snapshotted under the read lock before the supplier runs (the supplier still runs unlocked, so `Get` stays non-blocking).
- chore(deps): transitive bumps of `klauspost/compress`, `golang.org/x/net`, `golang.org/x/sys` and `golang.org/x/text`

## 1.3.6

- fix: a panicking discovery no longer publishes a successful empty result — the recovered panic is now surfaced as an error, so the cache records a failed discovery (and keeps returning an error) instead of an empty target/enrichment set with an advanced ETag

## 1.3.5

- fix: restore universal application of the discovery group attribute. v1.3.4 moved injection into the cache supplier chain, which skipped discoveries registered without `NewCachedTargetDiscovery` / `NewCachedEnrichmentDataDiscovery` (and any path going through `WithUpdate`). Injection now happens again in the HTTP discovery adapter, but each request builds a fresh copy of every target's attributes before adding `steadybit.group`. The discovery's underlying maps are never mutated, so this is safe under any level of concurrency and works for cached, non-cached, and incrementally-updated discoveries alike.

## 1.3.4

- fix (incomplete; superseded by 1.3.5): concurrent map iteration and map write when STEADYBIT_EXTENSION_DISCOVERY_GROUP is set. Group injection was moved into the cache supplier chain, fixing the crash but silently dropping the attribute for non-cached discoveries.

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

