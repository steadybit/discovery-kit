// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package discovery_kit_sdk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDiscoveryId = "com.steadybit.extension_kubernetes.kubernetes-deployment"

func target(name string, attributes map[string][]string) discovery_kit_api.Target {
	return discovery_kit_api.Target{
		Id:         name,
		Label:      name,
		TargetType: testDiscoveryId,
		Attributes: attributes,
	}
}

func names(targets []discovery_kit_api.Target) []string {
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		out = append(out, t.Id)
	}
	return out
}

func TestDiscoveryFilter_noConfigurationKeepsEverything(t *testing.T) {
	filter := newDiscoveryFilter(testDiscoveryId)
	require.Nil(t, filter, "no configuration should mean no filter at all, not an empty one")

	targets := []discovery_kit_api.Target{target("a", nil)}
	assert.Equal(t, targets, filter.retainTargets(targets), "a nil filter must be usable and keep everything")
}

func TestDiscoveryFilter_exclude(t *testing.T) {
	t.Setenv(excludeQueryEnvVar, `k8s.namespace IN (kube-system, istio-system)`)

	filter := newDiscoveryFilter(testDiscoveryId)
	require.NotNil(t, filter)

	retained := filter.retainTargets([]discovery_kit_api.Target{
		target("shop", map[string][]string{"k8s.namespace": {"shop"}}),
		target("coredns", map[string][]string{"k8s.namespace": {"kube-system"}}),
		target("istiod", map[string][]string{"k8s.namespace": {"istio-system"}}),
		target("no-namespace", nil),
	})

	assert.Equal(t, []string{"shop", "no-namespace"}, names(retained))
}

func TestDiscoveryFilter_include(t *testing.T) {
	t.Setenv(includeQueryEnvVar, `k8s.namespace="shop"`)

	retained := newDiscoveryFilter(testDiscoveryId).retainTargets([]discovery_kit_api.Target{
		target("shop", map[string][]string{"k8s.namespace": {"shop"}}),
		target("coredns", map[string][]string{"k8s.namespace": {"kube-system"}}),
		target("no-namespace", nil),
	})

	assert.Equal(t, []string{"shop"}, names(retained), "include drops everything it does not match")
}

func TestDiscoveryFilter_includeAndExcludeCombine(t *testing.T) {
	t.Setenv(includeQueryEnvVar, `k8s.namespace~"prod"`)
	t.Setenv(excludeQueryEnvVar, `k8s.label.tier="canary"`)

	retained := newDiscoveryFilter(testDiscoveryId).retainTargets([]discovery_kit_api.Target{
		target("prod-stable", map[string][]string{"k8s.namespace": {"prod-eu"}, "k8s.label.tier": {"stable"}}),
		target("prod-canary", map[string][]string{"k8s.namespace": {"prod-eu"}, "k8s.label.tier": {"canary"}}),
		target("dev-stable", map[string][]string{"k8s.namespace": {"dev"}, "k8s.label.tier": {"stable"}}),
	})

	assert.Equal(t, []string{"prod-stable"}, names(retained))
}

// target.type and steadybit.label are attributes on every target once the agent has seen it, but
// the agent has not run yet here. The filter supplies them so a query means the same thing on both
// sides of the agent.
func TestDiscoveryFilter_syntheticAttributes(t *testing.T) {
	t.Run("target.type", func(t *testing.T) {
		t.Setenv(excludeQueryEnvVar, `target.type="`+testDiscoveryId+`"`)
		retained := newDiscoveryFilter(testDiscoveryId).retainTargets([]discovery_kit_api.Target{target("a", nil)})
		assert.Empty(t, retained, "target.type must be resolvable even though it is not in Attributes")
	})

	t.Run("steadybit.label", func(t *testing.T) {
		t.Setenv(excludeQueryEnvVar, `steadybit.label~"temp-"`)
		retained := newDiscoveryFilter(testDiscoveryId).retainTargets([]discovery_kit_api.Target{
			target("temp-worker", nil),
			target("worker", nil),
		})
		assert.Equal(t, []string{"worker"}, names(retained))
	})

	t.Run("a real attribute still wins nothing weird", func(t *testing.T) {
		t.Setenv(excludeQueryEnvVar, `k8s.namespace="x"`)
		retained := newDiscoveryFilter(testDiscoveryId).retainTargets([]discovery_kit_api.Target{
			target("a", map[string][]string{"k8s.namespace": {"x"}}),
			target("b", map[string][]string{"k8s.namespace": {"y"}}),
		})
		assert.Equal(t, []string{"b"}, names(retained))
	})
}

func TestDiscoveryFilter_enrichmentData(t *testing.T) {
	t.Setenv(excludeQueryEnvVar, `source="legacy"`)

	retained := newDiscoveryFilter(testDiscoveryId).retainEnrichmentData([]discovery_kit_api.EnrichmentData{
		{Id: "keep", EnrichmentDataType: "t", Attributes: map[string][]string{"source": {"current"}}},
		{Id: "drop", EnrichmentDataType: "t", Attributes: map[string][]string{"source": {"legacy"}}},
	})

	require.Len(t, retained, 1)
	assert.Equal(t, "keep", retained[0].Id)
}

func TestDiscoveryFilter_enrichmentDataTargetType(t *testing.T) {
	t.Setenv(excludeQueryEnvVar, `target.type="enrichment-type"`)

	retained := newDiscoveryFilter(testDiscoveryId).retainEnrichmentData([]discovery_kit_api.EnrichmentData{
		{Id: "drop", EnrichmentDataType: "enrichment-type"},
		{Id: "keep", EnrichmentDataType: "other"},
	})

	require.Len(t, retained, 1)
	assert.Equal(t, "keep", retained[0].Id)
}

// There is no per-discovery configuration. Narrowing a query to one discovery of a multi-discovery
// extension is done in the query itself, which needs no second mechanism and no mapping of
// discovery ids onto environment variable names.
func TestDiscoveryFilter_scopingToOneDiscoveryIsDoneInTheQuery(t *testing.T) {
	t.Setenv(excludeQueryEnvVar, `target.type="`+testDiscoveryId+`" AND k8s.namespace="kube-system"`)

	filter := newDiscoveryFilter(testDiscoveryId)

	deployment := target("deployment", map[string][]string{"k8s.namespace": {"kube-system"}})
	pod := discovery_kit_api.Target{
		Id:         "pod",
		Label:      "pod",
		TargetType: "com.steadybit.extension_kubernetes.kubernetes-pod",
		Attributes: map[string][]string{"k8s.namespace": {"kube-system"}},
	}

	assert.Empty(t, names(filter.retainTargets([]discovery_kit_api.Target{deployment})),
		"the deployment matches both clauses and is excluded")
	assert.Equal(t, []string{"pod"}, names(filter.retainTargets([]discovery_kit_api.Target{pod})),
		"the same filter leaves another discovery's targets alone")
}

func TestDiscoveryFilter_queryFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "exclude.query")
	require.NoError(t, os.WriteFile(path, []byte("k8s.namespace=\"from-file\"\n"), 0600))
	t.Setenv(excludeQueryEnvVar+"_FILE", path)

	retained := newDiscoveryFilter(testDiscoveryId).retainTargets([]discovery_kit_api.Target{
		target("a", map[string][]string{"k8s.namespace": {"from-file"}}),
		target("b", map[string][]string{"k8s.namespace": {"other"}}),
	})

	assert.Equal(t, []string{"b"}, names(retained))
}

func TestDiscoveryFilter_literalWinsOverFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "exclude.query")
	require.NoError(t, os.WriteFile(path, []byte(`k8s.namespace="from-file"`), 0600))
	t.Setenv(excludeQueryEnvVar, `k8s.namespace="from-env"`)
	t.Setenv(excludeQueryEnvVar+"_FILE", path)

	retained := newDiscoveryFilter(testDiscoveryId).retainTargets([]discovery_kit_api.Target{
		target("env", map[string][]string{"k8s.namespace": {"from-env"}}),
		target("file", map[string][]string{"k8s.namespace": {"from-file"}}),
	})

	assert.Equal(t, []string{"file"}, names(retained))
}

// A silently ignored exclusion is worse than a failure to start: the operator believes targets are
// being withheld when they are not.
func TestDiscoveryFilter_malformedQueryPanics(t *testing.T) {
	t.Setenv(excludeQueryEnvVar, "k8s.namespace=")

	// Asserting on the parts we own, not on ANTLR's exact wording, which changes between versions.
	// The operator has to be able to see which variable was wrong, where, and what it contained.
	defer func() {
		recovered := recover()
		require.NotNil(t, recovered, "a malformed query must stop the extension, not be ignored")
		message, ok := recovered.(string)
		require.True(t, ok, "panic value should be a string, got %T", recovered)
		assert.Contains(t, message, excludeQueryEnvVar, "names the offending variable")
		assert.Contains(t, message, "line 1 column 14", "names the position")
		assert.Contains(t, message, "k8s.namespace=", "quotes the query")
	}()

	newDiscoveryFilter(testDiscoveryId)
}

func TestDiscoveryFilter_platformMarkerIsRejected(t *testing.T) {
	t.Setenv(excludeQueryEnvVar, "k8s.deployment={{deployment}}")

	assert.Panics(t, func() { newDiscoveryFilter(testDiscoveryId) },
		"markers resolve against platform state an extension cannot see")
}

func TestDiscoveryFilter_unreadableFilePanics(t *testing.T) {
	t.Setenv(excludeQueryEnvVar+"_FILE", filepath.Join(t.TempDir(), "does-not-exist"))

	assert.Panics(t, func() { newDiscoveryFilter(testDiscoveryId) })
}

func TestDiscoveryFilter_blankQueryIsNotAFilter(t *testing.T) {
	t.Setenv(excludeQueryEnvVar, "   ")
	t.Setenv(includeQueryEnvVar, "")

	assert.Nil(t, newDiscoveryFilter(testDiscoveryId),
		"an empty query means match-all, which as a filter means no filter — not 'drop everything'")
}
