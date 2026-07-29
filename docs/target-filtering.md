# Target Filtering

An extension can be told not to report certain targets at all. The filter is configured by whoever
operates the extension, using the same query language as environments and experiment blast radii.

This is different from the `steadybit.com/discovery-disabled` label some extensions support. That
label is set by whoever owns the *resource*; this filter is set by whoever owns the *extension*, and
needs no cooperation from the teams whose namespaces or accounts are being discovered.

## Why filter in the extension

Filtering here rather than in the platform means an excluded target is never serialized into a
discovery response, never transferred to the agent, and never ingested or stored. It costs nothing
downstream because it never exists downstream.

The usual reason to reach for it is volume: a cluster or account containing large numbers of targets
nobody will ever run an experiment against.

## Configuration

| Environment Variable                          | Meaning                                          |
|-----------------------------------------------|--------------------------------------------------|
| `STEADYBIT_EXTENSION_DISCOVERY_EXCLUDE_QUERY` | targets matching the query are **not** reported   |
| `STEADYBIT_EXTENSION_DISCOVERY_INCLUDE_QUERY` | **only** targets matching the query are reported  |

```shell
STEADYBIT_EXTENSION_DISCOVERY_EXCLUDE_QUERY='k8s.namespace IN (kube-system, istio-system)'
```

Both are optional and combine: a target is reported when it matches the include query (if set) and
does not match the exclude query (if set).

Either variable also accepts a `_FILE` suffix, reading the query from a file instead — for queries
too long for an environment variable, or supplied through a ConfigMap or secret:

```shell
STEADYBIT_EXTENSION_DISCOVERY_EXCLUDE_QUERY_FILE=/etc/steadybit/exclude.query
```

A query applies to **every discovery of the extension**. To narrow it to one discovery of an
extension that hosts several, say so in the query rather than looking for a second setting:

```shell
STEADYBIT_EXTENSION_DISCOVERY_EXCLUDE_QUERY='target.type="com.steadybit.extension_kubernetes.kubernetes-pod" AND k8s.namespace="kube-system"'
```

## Writing the query

The syntax is Steadybit's target query language — the same one the target explorer and environment
editor use, so a query can be drafted there and pasted into the configuration:

```
k8s.namespace="prod"
k8s.namespace IN (kube-system, istio-system)
k8s.label.tier!="canary" AND k8s.namespace~"prod"
NOT (aws.account="123456789012" OR aws.zone.id="euc1-az1")
container.image~"registry.internal/"
count(k8s.container.name)>3
host.hostname IS NOT PRESENT
```

Operators: `=` `!=` `~` (contains) `!~`, each with a `*` suffix for the case-insensitive form
(`=*`, `~*`, …); `IN (…)` / `NOT IN (…)`; `IS PRESENT` / `IS NOT PRESENT`; `count(key)` with
`= != > >= < <=`. Combine with `AND`, `OR`, `NOT` and parentheses — `NOT` binds tighter than `AND`,
which binds tighter than `OR`.

Two attributes are usable even though a discovery does not put them in a target's attribute map:

- **`target.type`** — the target type id, as in the examples above.
- **`steadybit.label`** — the target's label. See
  [Reserved Target Attributes](/docs/reserved-target-attributes.md).

Both are ordinary attributes once the agent has processed a target; the filter supplies them early
so that a query selects the same targets in the extension as it would in the target explorer.

Enrichment data is filtered by the same queries, with `target.type` matching its enrichment data
type.

## Behaviour worth knowing

- **A malformed query stops the extension at start up.** An exclusion an operator believes is in
  force but which silently does nothing is worse than a failure to start, so the query is parsed
  once at start up and a parse error is fatal, naming the variable, the position and the query.
- **The active filter is logged at start up** (`INFO`), and the number of dropped records per
  discovery at `DEBUG`. Without that, a missing target is undebuggable — the extension simply never
  mentions it.
- **An empty or unset query is not a filter.** It does not mean "match nothing".
- **Variable (`{{...}}`) and template-placeholder (`[[...]]`) markers are rejected.** The platform
  resolves those against experiment and environment state that an extension cannot see.
- **Filtering is not retroactive.** Targets already reported stay in the platform until the agent
  next sends a discovery result that omits them.
- **Excluded targets are invisible, not merely hidden.** They cannot be selected in an experiment,
  they do not appear in the target explorer, and advice will not be computed for them.

## Availability

The filter is implemented in the Go
[`discovery_kit_sdk`](/go/discovery_kit_sdk/README.md), from v1.4.1. Extensions that serve discovery
endpoints without the SDK are unaffected and would have to apply a filter themselves.
