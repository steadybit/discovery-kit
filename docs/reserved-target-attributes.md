# Reserved Target Attributes

Some target attributes have a special meaning within Steadybit. They are reserved: extensions and operators should only set them as described below, since the platform, the agent, and the enrichment rules rely on their semantics.

## `steadybit.label`

Used whenever the Steadybit UI needs to render a label for a target. The agent sets this attribute automatically from the target's `Label` field. Extensions may also set it explicitly in discovery code, but setting the `Label` field is prefered.

## `steadybit.extension`

Added by the agent and identifies the specific extension instance that reported the target. The platform uses it for action routing — i.e. to send action requests back to the same extension that discovered the target. Extensions must not set this attribute themselves.

## `steadybit.group`

Defaults to `default` (set by the agent). Can be overridden per extension by setting the environment variable `STEADYBIT_EXTENSION_GROUP`; the discovery-kit SDK then injects the value into every target and enrichment data record reported by that extension instance.

All enrichment rules include `steadybit.group` as an additional matcher. This lets you slice targets along a dimension that is not part of the discovered data — for example, to keep environments separate when the same Steadybit installation discovers targets from multiple stages or tenants, and to prevent the enrichment rules from matching across those groups.
