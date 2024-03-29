// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_api

import (
	"testing"
)

// Note: These test cases only check that the code compiles as intended.
// On compilation errors, we most likely have caused a breaking change of
// the API and need to act accordingly.

func markAsUsed(t *testing.T, v any) {
	if v == nil {
		t.Fail()
	}
}

func TestDiscoveryList(t *testing.T) {
	v := DiscoveryList{
		Discoveries: []DescribingEndpointReference{
			{
				Method: "GET",
				Path:   "/",
			},
		},
		TargetTypes: []DescribingEndpointReference{
			{
				Method: "GET",
				Path:   "/",
			},
		},
		TargetAttributes: []DescribingEndpointReference{
			{
				Method: "GET",
				Path:   "/",
			},
		},
	}
	markAsUsed(t, v)
}

func TestDiscoveryDescription(t *testing.T) {
	v := DiscoveryDescription{
		Discover: DescribingEndpointReferenceWithCallInterval{
			Method:       "GET",
			Path:         "/",
			CallInterval: Ptr("5m"),
		},
		Id: "42",
	}
	markAsUsed(t, v)
}

func TestAttributeDescriptions(t *testing.T) {
	v := AttributeDescriptions{
		Attributes: []AttributeDescription{
			{
				Attribute: "k8s.deployment",
				Label: PluralLabel{
					One:   "Kubernetes deployment",
					Other: "Kubernetes deployments",
				},
			},
		},
	}
	markAsUsed(t, v)
}

func TestDiscoveredTargets(t *testing.T) {
	v := DiscoveredTargets{
		Targets: []Target{
			{
				Attributes: make(map[string][]string),
				Id:         "i",
				Label:      "l",
				TargetType: "t",
			},
		},
	}
	markAsUsed(t, v)
}

func TestDiscoveryData(t *testing.T) {
	v := DiscoveryData{
		Targets: &[]Target{
			{
				Attributes: make(map[string][]string),
				Id:         "i",
				Label:      "l",
				TargetType: "t",
			},
		},
		EnrichmentData: &[]EnrichmentData{
			{
				Attributes:         make(map[string][]string),
				Id:                 "i",
				EnrichmentDataType: "t",
			},
		},
	}
	markAsUsed(t, v)
}

func TestTargetDescription(t *testing.T) {
	v := TargetDescription{
		Category: Ptr("basic"),
		Icon:     Ptr("data:..."),
		Id:       "id",
		Version:  "1.0.0",
		Label: PluralLabel{
			One:   "one",
			Other: "other",
		},
		Table: Table{
			Columns: []Column{
				{
					Attribute:          "attr",
					FallbackAttributes: Ptr([]string{"a", "b"}),
				},
			},
			OrderBy: []OrderBy{
				{
					Attribute: "attr",
					Direction: DESC,
				},
			},
		},
	}
	markAsUsed(t, v)
}

func TestDiscoveryKitError(t *testing.T) {
	v := DiscoveryKitError{
		Detail:   Ptr("d"),
		Instance: Ptr("i"),
		Title:    "t",
		Type:     Ptr("t"),
	}
	markAsUsed(t, v)
}
