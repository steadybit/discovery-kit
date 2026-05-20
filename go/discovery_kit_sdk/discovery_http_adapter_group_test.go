// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package discovery_kit_sdk

import (
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func Test_withGroupAttribute_copies_map(t *testing.T) {
	src := map[string][]string{"k": {"v"}}
	targets := []discovery_kit_api.Target{{Id: "a", Attributes: src}}

	out := withGroupAttribute(targets, "prod-eu")

	assert.Equal(t, []string{"prod-eu"}, out[0].Attributes[groupAttributeKey])
	assert.Equal(t, []string{"v"}, out[0].Attributes["k"])
	// original map is untouched
	_, present := src[groupAttributeKey]
	assert.False(t, present, "must not mutate the input map")
	// returned map is a different allocation
	assert.NotSame(t, &targets[0].Attributes, &out[0].Attributes)
}

func Test_withGroupAttributeEnrichment_copies_map(t *testing.T) {
	src := map[string][]string{"k": {"v"}}
	data := []discovery_kit_api.EnrichmentData{{Id: "a", Attributes: src}}

	out := withGroupAttributeEnrichment(data, "prod-eu")

	assert.Equal(t, []string{"prod-eu"}, out[0].Attributes[groupAttributeKey])
	_, present := src[groupAttributeKey]
	assert.False(t, present, "must not mutate the input map")
}

// Test_withGroupAttribute_is_concurrent_safe locks in the fix for the
// "concurrent map iteration and map write" panic. Many goroutines call
// withGroupAttribute on a shared source slice while other goroutines iterate
// the returned maps — the source maps must never be written to, so iteration
// against them stays race-free even with the race detector enabled.
func Test_withGroupAttribute_is_concurrent_safe(t *testing.T) {
	shared := []discovery_kit_api.Target{
		{Id: "a", Attributes: map[string][]string{"k": {"v"}}},
		{Id: "b", Attributes: map[string][]string{"k": {"v"}}},
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				out := withGroupAttribute(shared, "prod-eu")
				for _, tg := range out {
					for k, v := range tg.Attributes {
						_ = k
						_ = v
					}
				}
				// also iterate the shared source as if encoding it for a JSON response
				for _, tg := range shared {
					for k, v := range tg.Attributes {
						_ = k
						_ = v
					}
				}
			}
		}()
	}
	wg.Wait()
}

func Test_copyAttributesWithGroup_handles_nil_source(t *testing.T) {
	out := copyAttributesWithGroup(nil, "g")
	assert.Equal(t, []string{"g"}, out[groupAttributeKey])
}
