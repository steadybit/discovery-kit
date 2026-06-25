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

	out := normalizeTargets(targets, "prod-eu")

	assert.Equal(t, []string{"prod-eu"}, out[0].Attributes[groupAttributeKey])
	assert.Equal(t, []string{"v"}, out[0].Attributes["k"])
	// original map is untouched
	_, present := src[groupAttributeKey]
	assert.False(t, present, "must not mutate the input map")
	// returned map is a different allocation
	assert.NotSame(t, &targets[0].Attributes, &out[0].Attributes)
}

func Test_normalizeEnrichmentData_copies_map(t *testing.T) {
	src := map[string][]string{"k": {"v"}}
	data := []discovery_kit_api.EnrichmentData{{Id: "a", Attributes: src}}

	out := normalizeEnrichmentData(data, "prod-eu")

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
				out := normalizeTargets(shared, "prod-eu")
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

func Test_normalizeAttributes_handles_nil_source(t *testing.T) {
	out := normalizeAttributes(nil, "g")
	assert.Equal(t, []string{"g"}, out[groupAttributeKey])
}

// Test_normalizeAttributes_sorts_multivalued locks in the fix for the
// extension-kubernetes platform-DB churn incident: multi-valued attribute slices
// must come out in a stable order even when the source is in random Go-map order,
// otherwise the platform's target-diff sees a change every cycle and re-writes.
func Test_normalizeAttributes_sorts_multivalued(t *testing.T) {
	src := map[string][]string{
		"single":  {"only"},
		"k8s.hpa": {"bbb", "aaa", "ccc"},
		"k8s.pdb": {"zzz", "aaa"},
	}
	out := normalizeAttributes(src, "")

	assert.Equal(t, []string{"only"}, out["single"], "single-valued slice untouched")
	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, out["k8s.hpa"], "multi-valued slice sorted")
	assert.Equal(t, []string{"aaa", "zzz"}, out["k8s.pdb"], "multi-valued slice sorted")
	// Source must not be mutated.
	assert.Equal(t, []string{"bbb", "aaa", "ccc"}, src["k8s.hpa"], "source slice must not be mutated")
}
