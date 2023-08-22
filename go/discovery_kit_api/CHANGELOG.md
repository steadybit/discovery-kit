# Changelog

## 1.4.1
- Breaking: `enrichmentRules` are no longer an attribute of TargetDescription. Instead, they are now a first level entity as part of `DiscoveryList`. As part of the move to an own entity, `TargetEnrichmentRule` needs an `id` and a `version` attribute.
- Discoveries can now return `enrichmentData` besides `targets`

## 1.4.0

- please use 1.4.1 instead

## 1.3.0

- Removed restriction of discoveries to AWS agents.

## 1.2.0

- Support target enrichment rules.

## 1.0.0

 - Empty release just to bump the version number to 1.0.0.

## 0.1.0

 - Initial release