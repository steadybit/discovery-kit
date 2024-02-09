# Changelog

## 1.5.2
- add TargetEnrichmentRule Matcher Regex (available in platform version >= 2.0.0 and agent version >= 2.0.2)

## 1.5.1
- Deprecate: DiscoveryDescription.restrictTo. Not needed anymore.

## 1.5.0
- Target Ids must be unique across target type
- Aligned Http Method constants

## 1.4.2
- Embed the openapi spec into the api package

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