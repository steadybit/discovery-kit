// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/exthttp"
	"time"
)

var (
	registeredDiscoveries                  = make(map[string]Discovery)
	registeredTargetDescriber              = make(map[string]TargetDescriber)
	registeredAttributeDescriber           = make([]AttributeDescriber, 0)
	registeredEnrichmentRulesContributions = make(map[string]discovery_kit_api.TargetEnrichmentRule)
)

type Discovery interface {
	// Describe returns the discovery description.
	Describe() discovery_kit_api.DiscoveryDescription
}

type TargetDiscovery interface {
	Discovery
	// DiscoverTargets returns a list of targets.
	DiscoverTargets(ctx context.Context) ([]discovery_kit_api.Target, error)
}

type EnrichmentDataDiscovery interface {
	Discovery
	// DiscoverEnrichmentData returns a list of enrichment data.
	DiscoverEnrichmentData(ctx context.Context) ([]discovery_kit_api.EnrichmentData, error)
}

type TargetDescriber interface {
	// DescribeTarget returns the target description.
	DescribeTarget() discovery_kit_api.TargetDescription
}

type AttributeDescriber interface {
	// DescribeAttributes returns the target attribute description.
	DescribeAttributes() []discovery_kit_api.AttributeDescription
}

type EnrichmentRulesDescriber interface {
	// DescribeEnrichmentRules returns a list of target enrichment rules.
	DescribeEnrichmentRules() []discovery_kit_api.TargetEnrichmentRule
}

type p interface {
	LastModified() time.Time
}

type Unwrapper interface {
	Unwrap() interface{}
}

func Register(o interface{}) {
	matched := false
	matched = registerDiscovery(o) || matched
	matched = registerTargetDescriber(o) || matched
	matched = registerAttributeDescriber(o) || matched
	matched = registerEnrichmentRuleContributor(o) || matched

	if !matched {
		panic(fmt.Sprintf("unknown discovery type: %T", o))
	}
}

func registerDiscovery(o interface{}) bool {
	if d, ok := o.(Discovery); ok {
		adapter := newDiscoveryHttpAdapter(d)
		registeredDiscoveries[adapter.description.Id] = d
		adapter.registerHandlers()
		return true
	}
	if w, ok := o.(Unwrapper); ok {
		return registerDiscovery(w.Unwrap())
	}
	return false
}

func registerTargetDescriber(o interface{}) bool {
	if d, ok := o.(TargetDescriber); ok {
		id := d.DescribeTarget().Id
		registeredTargetDescriber[id] = d
		exthttp.RegisterHttpHandler(fmt.Sprintf("/%s/discovery/target-description", id), exthttp.GetterAsHandler(d.DescribeTarget))
		return true
	}
	if w, ok := o.(Unwrapper); ok {
		return registerTargetDescriber(w.Unwrap())
	}
	return false
}

func registerAttributeDescriber(o interface{}) bool {
	if d, ok := o.(AttributeDescriber); ok {
		if len(registeredAttributeDescriber) == 0 {
			exthttp.RegisterHttpHandler("/discovery/attributes", exthttp.GetterAsHandler(describeAttributes))
		}
		registeredAttributeDescriber = append(registeredAttributeDescriber, d)
		return true
	}
	if w, ok := o.(Unwrapper); ok {
		return registerAttributeDescriber(w.Unwrap())
	}
	return false
}

func registerEnrichmentRuleContributor(o interface{}) bool {
	if d, ok := o.(EnrichmentRulesDescriber); ok {
		for _, rule := range d.DescribeEnrichmentRules() {
			ruleCopy := rule // copy the value, otherwise the closure will always point to the last value of the slice
			registeredEnrichmentRulesContributions[rule.Id] = rule
			exthttp.RegisterHttpHandler(fmt.Sprintf("/discovery/enrichment-rules/%s", rule.Id), exthttp.GetterAsHandler(func() discovery_kit_api.TargetEnrichmentRule {
				return ruleCopy
			}))
		}
		return true
	}
	if w, ok := o.(Unwrapper); ok {
		return registerEnrichmentRuleContributor(w.Unwrap())
	}
	return false
}

func GetDiscoveryList() discovery_kit_api.DiscoveryList {
	checkForDuplicates(mergeAllAttributeDescriptions())

	return discovery_kit_api.DiscoveryList{
		Discoveries:           getDiscoveryReferences(),
		TargetAttributes:      getTargetAttributeReferences(),
		TargetEnrichmentRules: getTargetEnrichmentRuleReferences(),
		TargetTypes:           getTargetTypeReferences(),
	}
}

func getDiscoveryReferences() []discovery_kit_api.DescribingEndpointReference {
	result := make([]discovery_kit_api.DescribingEndpointReference, 0, len(registeredDiscoveries))
	for id := range registeredDiscoveries {
		result = append(result, discovery_kit_api.DescribingEndpointReference{
			Method: discovery_kit_api.GET,
			Path:   fmt.Sprintf("/%s/discovery", id),
		})
	}
	return result
}
func getTargetAttributeReferences() []discovery_kit_api.DescribingEndpointReference {
	if len(registeredAttributeDescriber) == 0 {
		return []discovery_kit_api.DescribingEndpointReference{}
	}

	return []discovery_kit_api.DescribingEndpointReference{{
		Method: discovery_kit_api.GET,
		Path:   "/discovery/attributes",
	}}
}

func getTargetEnrichmentRuleReferences() []discovery_kit_api.DescribingEndpointReference {
	result := make([]discovery_kit_api.DescribingEndpointReference, 0, len(registeredEnrichmentRulesContributions))
	for id := range registeredEnrichmentRulesContributions {
		result = append(result, discovery_kit_api.DescribingEndpointReference{
			Method: discovery_kit_api.GET,
			Path:   fmt.Sprintf("/discovery/enrichment-rules/%s", id),
		})
	}
	return result
}

func getTargetTypeReferences() []discovery_kit_api.DescribingEndpointReference {
	result := make([]discovery_kit_api.DescribingEndpointReference, 0, len(registeredTargetDescriber))
	for id := range registeredTargetDescriber {
		result = append(result, discovery_kit_api.DescribingEndpointReference{
			Method: discovery_kit_api.GET,
			Path:   fmt.Sprintf("/%s/discovery/target-description", id),
		})
	}
	return result
}

func describeAttributes() discovery_kit_api.AttributeDescriptions {
	return discovery_kit_api.AttributeDescriptions{Attributes: mergeAllAttributeDescriptions()}
}

func mergeAllAttributeDescriptions() []discovery_kit_api.AttributeDescription {
	var result []discovery_kit_api.AttributeDescription
	for _, describer := range registeredAttributeDescriber {
		descriptions := describer.DescribeAttributes()
		result = append(result, descriptions...)
	}
	return result
}

func checkForDuplicates(descriptions []discovery_kit_api.AttributeDescription) {
	var duplicateCheck = make(map[string]int)

	for _, description := range descriptions {
		duplicateCheck[description.Attribute] = duplicateCheck[description.Attribute] + 1
	}

	for attribute, count := range duplicateCheck {
		if count > 1 {
			log.Warn().Int("count", count).Msgf("attribute %s is defined multiple times", attribute)
		}
	}
}

func ClearRegisteredDiscoveries() {
	registeredDiscoveries = make(map[string]Discovery)
	registeredTargetDescriber = make(map[string]TargetDescriber)
	registeredAttributeDescriber = make([]AttributeDescriber, 0)
	registeredEnrichmentRulesContributions = make(map[string]discovery_kit_api.TargetEnrichmentRule)
}
