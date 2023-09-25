package discovery_kit_commons

import (
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"strings"
)

func ApplyAttributeExcludes(targets []discovery_kit_api.Target, denyList []string) []discovery_kit_api.Target {
	if denyList == nil || (denyList != nil && len(denyList) == 0) {
		return targets
	}
	resultTargets := make([]discovery_kit_api.Target, len(targets)) // we do not want to modify the original targets
	for i, target := range targets {
		resultTargets[i] = target
		resultTargets[i].Attributes = applyExcludesToAttributes(target.Attributes, denyList)
	}
	return resultTargets
}

func ApplyAttributeExcludesToEnrichmentData(targets []discovery_kit_api.EnrichmentData, denyList []string) []discovery_kit_api.EnrichmentData {
	if denyList == nil || (denyList != nil && len(denyList) == 0) {
		return targets
	}
	resultTargets := make([]discovery_kit_api.EnrichmentData, len(targets)) // we do not want to modify the original enrichment data
	for i, target := range targets {
		resultTargets[i] = target
		resultTargets[i].Attributes = applyExcludesToAttributes(target.Attributes, denyList)
	}
	return resultTargets
}

func applyExcludesToAttributes(attributes map[string][]string, denyList []string) map[string][]string {
	resultAttributes := make(map[string][]string) // we do not want to modify the original attributes
	for key := range attributes {
		resultAttributes[key] = attributes[key]
		for _, denyListEntry := range denyList {
			if strings.HasSuffix(denyListEntry, "*") {
				// if the deny list entry ends with a wildcard, check if the key starts with the deny list entry
				if strings.HasPrefix(key, strings.TrimSuffix(denyListEntry, "*")) {
					delete(resultAttributes, key)
				}
			} else {
				// if the deny list entry does not end with a wildcard, check if the key is equal to the deny list entry
				if key == denyListEntry {
					delete(resultAttributes, key)
				}
			}
		}
	}
	return resultAttributes
}
