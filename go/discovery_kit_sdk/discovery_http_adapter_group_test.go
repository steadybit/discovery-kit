// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package discovery_kit_sdk

import (
	"sync"
	"testing"

	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_normalizeTargets_copies_map_when_group_set(t *testing.T) {
	src := map[string][]string{"k": {"v"}}
	targets := []discovery_kit_api.Target{{Id: "a", Attributes: src}}

	out := normalizeTargets(targets, "prod-eu")

	assert.Equal(t, []string{"prod-eu"}, out[0].Attributes[groupAttributeKey])
	assert.Equal(t, []string{"v"}, out[0].Attributes["k"])
	// Source must not be mutated by group injection.
	_, present := src[groupAttributeKey]
	assert.False(t, present, "must not mutate the input map")
	// Mutating the returned map must not be visible in the source — proves
	// they are different map allocations regardless of internal aliasing of
	// individual slice pointers.
	out[0].Attributes["sentinel"] = []string{"x"}
	_, leaked := src["sentinel"]
	assert.False(t, leaked, "returned map must be a different allocation than the source")
}

func Test_normalizeEnrichmentData_copies_map_when_group_set(t *testing.T) {
	src := map[string][]string{"k": {"v"}}
	data := []discovery_kit_api.EnrichmentData{{Id: "a", Attributes: src}}

	out := normalizeEnrichmentData(data, "prod-eu")

	assert.Equal(t, []string{"prod-eu"}, out[0].Attributes[groupAttributeKey])
	_, present := src[groupAttributeKey]
	assert.False(t, present, "must not mutate the input map")
	out[0].Attributes["sentinel"] = []string{"x"}
	_, leaked := src["sentinel"]
	assert.False(t, leaked, "returned map must be a different allocation than the source")
}

// Test_normalizeTargets_is_concurrent_safe locks in the fix for the
// "concurrent map iteration and map write" panic plus the multi-valued sort
// path. Many goroutines call normalizeTargets on a shared source slice that
// contains both single- and multi-valued attributes (the latter intentionally
// unsorted so the sort branch is exercised), while other goroutines iterate
// the source as if encoding it for a JSON response — the source maps must
// never be written to, so the loop stays race-free even with -race.
func Test_normalizeTargets_is_concurrent_safe(t *testing.T) {
	shared := []discovery_kit_api.Target{
		{Id: "a", Attributes: map[string][]string{"k": {"v"}, "k8s.hpa": {"bbb", "aaa", "ccc"}}},
		{Id: "b", Attributes: map[string][]string{"k": {"v"}, "k8s.pdb": {"zzz", "aaa"}}},
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
	// Source slices must still be in their original order after the concurrent run.
	assert.Equal(t, []string{"bbb", "aaa", "ccc"}, shared[0].Attributes["k8s.hpa"], "source slice mutated by normalize")
	assert.Equal(t, []string{"zzz", "aaa"}, shared[1].Attributes["k8s.pdb"], "source slice mutated by normalize")
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
	// group=="" must NOT inject a steadybit.group attribute — without this assertion a regression
	// that drops the `if group != ""` guard would silently ship phantom `steadybit.group: [""]`
	// attributes on every target.
	_, hasGroup := out[groupAttributeKey]
	assert.False(t, hasGroup, "group key must be absent when group is empty")
	// Both source slices must be untouched (catches an in-place-sort regression
	// even when one attribute happens to be 2 elements — the case most likely
	// to be "optimised" into a swap-in-place).
	assert.Equal(t, []string{"bbb", "aaa", "ccc"}, src["k8s.hpa"], "k8s.hpa source slice must not be mutated")
	assert.Equal(t, []string{"zzz", "aaa"}, src["k8s.pdb"], "k8s.pdb source slice must not be mutated")
}

// Test_normalizeAttributes_sorts_multivalued_with_group covers the most
// production-relevant combination: STEADYBIT_EXTENSION_DISCOVERY_GROUP is set
// AND the extension reports multi-valued attributes. A future change that
// only sorts when group=="" (e.g. gated on extra==0) would pass every other
// test in this file and ship the original platform-DB churn bug to every
// customer running with a group.
func Test_normalizeAttributes_sorts_multivalued_with_group(t *testing.T) {
	src := map[string][]string{
		"k8s.hpa": {"bbb", "aaa", "ccc"},
	}
	out := normalizeAttributes(src, "prod-eu")

	assert.Equal(t, []string{"prod-eu"}, out[groupAttributeKey], "group attribute set")
	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, out["k8s.hpa"], "multi-valued sorted even when group is set")
	assert.Equal(t, []string{"bbb", "aaa", "ccc"}, src["k8s.hpa"], "source slice must not be mutated")
}

// Test_normalizeTargets_fast_path_no_copy_when_already_sorted verifies that
// targets whose multi-valued attributes are already sorted and that have no
// group configured are returned without a fresh allocation. Avoids a
// per-cycle map-copy regression on large-fan-out discoveries.
//
// The test asserts the *backing array is shared* (slice header identity), not
// caller-visible mutation through the boundary. A future maintainer who
// legitimately reintroduces a defensive copy (e.g. to harden against a new
// mutate-after-return contract) should be able to revisit this test by
// changing the assertion — not be tempted to revert their copy.
func Test_normalizeTargets_fast_path_no_copy_when_already_sorted(t *testing.T) {
	targets := []discovery_kit_api.Target{
		{Id: "a", Attributes: map[string][]string{"single": {"only"}, "k8s.hpa": {"aaa", "bbb", "ccc"}}},
		{Id: "b", Attributes: map[string][]string{"k8s.pdb": {"aaa", "zzz"}}},
	}

	out := normalizeTargets(targets, "")

	// Slice-header identity: same backing array, same len/cap. Not just element-0 aliasing —
	// a regression like `return targets[:1:1]` would satisfy &targets[0]==&out[0] but break
	// callers who rely on the full input being returned.
	require.Equal(t, len(targets), len(out), "fast path must return all targets")
	assert.Same(t, &targets[0], &out[0], "fast path must share the underlying array")
	assert.Same(t, &targets[1], &out[1], "every element must alias, not just the first")
}

// Test_normalizeEnrichmentData_fast_path_no_copy_when_already_sorted mirrors
// the targets fast-path test for enrichment data — without it, a future
// regression that inverts enrichmentNeedsNormalize would silently re-introduce
// per-cycle map allocations for enrichment-data discoveries.
func Test_normalizeEnrichmentData_fast_path_no_copy_when_already_sorted(t *testing.T) {
	data := []discovery_kit_api.EnrichmentData{
		{Id: "a", Attributes: map[string][]string{"single": {"only"}, "k8s.hpa": {"aaa", "bbb"}}},
		{Id: "b", Attributes: map[string][]string{"k8s.pdb": {"aaa", "zzz"}}},
	}

	out := normalizeEnrichmentData(data, "")

	require.Equal(t, len(data), len(out))
	assert.Same(t, &data[0], &out[0], "fast path must share the underlying array")
	assert.Same(t, &data[1], &out[1])
}

// Test_normalizeTargets_slow_path_allocates_when_unsorted covers the wiring
// from normalizeTargets through to normalizeAttributes: with no group and an
// unsorted multi-valued attribute on at least one target, the SDK MUST
// allocate a fresh map for that target with the slice sorted on the wire,
// and the source slice must remain untouched.
func Test_normalizeTargets_slow_path_allocates_when_unsorted(t *testing.T) {
	srcAttrs := map[string][]string{"k8s.hpa": {"bbb", "aaa"}} // unsorted
	targets := []discovery_kit_api.Target{{Id: "a", Attributes: srcAttrs}}

	out := normalizeTargets(targets, "")

	// Wire output must be sorted — this is the SDK's primary contract.
	assert.Equal(t, []string{"aaa", "bbb"}, out[0].Attributes["k8s.hpa"], "wire output sorted")
	// Source slice must be untouched.
	assert.Equal(t, []string{"bbb", "aaa"}, srcAttrs["k8s.hpa"], "source slice not mutated")
	// Mutating the returned Attributes map must NOT leak into the source — proves the slow path
	// produced a fresh map.
	out[0].Attributes["sentinel"] = []string{"x"}
	_, leaked := srcAttrs["sentinel"]
	assert.False(t, leaked, "slow path must return a fresh Attributes map")
}
