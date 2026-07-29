// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package discovery_kit_sdk

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/extquery"
)

const (
	excludeQueryEnvVar = "STEADYBIT_EXTENSION_DISCOVERY_EXCLUDE_QUERY"
	includeQueryEnvVar = "STEADYBIT_EXTENSION_DISCOVERY_INCLUDE_QUERY"

	targetTypeAttribute = "target.type"
	labelAttribute      = "steadybit.label"
)

// filterLogOnce keeps the "filter active" line to one occurrence: newDiscoveryFilter runs once per
// registered discovery, and the filter is identical for all of them.
var filterLogOnce sync.Once

// discoveryFilter drops targets an operator does not want reported, using the same query language
// the platform uses for environments and blast radii. Filtering here rather than in the platform
// means the targets never reach the agent at all: no transfer, no ingest, no rows.
//
// A nil filter keeps everything, which is the common case.
type discoveryFilter struct {
	discoveryId string
	include     extquery.Predicate
	exclude     extquery.Predicate
}

// newDiscoveryFilter resolves the filter, which applies to every discovery of the extension. Both
// queries also accept a _FILE suffix pointing at a file, since a query can outgrow an env var.
//
// There is deliberately no way to scope a query to a single discovery: a query can say
// `target.type="..."` itself, which does the same job without a second configuration mechanism and
// without mangling discovery ids into environment variable names.
//
// A malformed query panics rather than being ignored. An exclusion the operator believes is in
// force but which silently does nothing is worse than a failure to start, and this is caught the
// first time the extension runs.
func newDiscoveryFilter(discoveryId string) *discoveryFilter {
	include := mustParseQueryFromEnv(includeQueryEnvVar)
	exclude := mustParseQueryFromEnv(excludeQueryEnvVar)
	if include == nil && exclude == nil {
		return nil
	}

	// Called once per registered discovery, but the filter is the same for all of them — an
	// extension with ten discoveries would otherwise log ten identical lines at start up.
	filterLogOnce.Do(func() {
		log.Info().
			Str("include", predicateString(include)).
			Str("exclude", predicateString(exclude)).
			Msg("Discovery filter active. Targets not matching include, or matching exclude, are not reported.")
	})

	return &discoveryFilter{discoveryId: discoveryId, include: include, exclude: exclude}
}

func (f *discoveryFilter) retainTargets(targets []discovery_kit_api.Target) []discovery_kit_api.Target {
	if f == nil {
		return targets
	}
	retained := make([]discovery_kit_api.Target, 0, len(targets))
	for _, target := range targets {
		if f.keep(targetAttributes(target)) {
			retained = append(retained, target)
		}
	}
	f.logDropped(len(targets), len(retained), "targets")
	return retained
}

func (f *discoveryFilter) retainEnrichmentData(data []discovery_kit_api.EnrichmentData) []discovery_kit_api.EnrichmentData {
	if f == nil {
		return data
	}
	retained := make([]discovery_kit_api.EnrichmentData, 0, len(data))
	for _, d := range data {
		if f.keep(enrichmentDataAttributes(d)) {
			retained = append(retained, d)
		}
	}
	f.logDropped(len(data), len(retained), "enrichment data records")
	return retained
}

func (f *discoveryFilter) keep(attributes extquery.Attributes) bool {
	if f.include != nil && !f.include.Matches(attributes) {
		return false
	}
	return f.exclude == nil || !f.exclude.Matches(attributes)
}

// targetAttributes exposes a target to the query language. `target.type` and `steadybit.label` are
// ordinary attributes on every target in the platform, but the agent is what adds them and it has
// not run yet at this point — so they are supplied here. Without that, the same query text would
// select differently in an extension than in the target explorer.
func targetAttributes(target discovery_kit_api.Target) extquery.Attributes {
	return extquery.AttributesFunc(func(key string) []string {
		switch key {
		case targetTypeAttribute:
			return []string{target.TargetType}
		case labelAttribute:
			return []string{target.Label}
		}
		return target.Attributes[key]
	})
}

// enrichmentDataAttributes is the same for enrichment data, which has no label.
func enrichmentDataAttributes(data discovery_kit_api.EnrichmentData) extquery.Attributes {
	return extquery.AttributesFunc(func(key string) []string {
		if key == targetTypeAttribute {
			return []string{data.EnrichmentDataType}
		}
		return data.Attributes[key]
	})
}

func mustParseQueryFromEnv(envVar string) extquery.Predicate {
	query, source := queryFromEnv(envVar)
	if strings.TrimSpace(query) == "" {
		return nil
	}
	predicate, err := extquery.Parse(query)
	if err != nil {
		panic(fmt.Sprintf("discovery filter %s is not a valid query: %s (query: %s)", source, err, query))
	}
	return predicate
}

// queryFromEnv returns the query and the name of the variable it came from, preferring a literal
// value over a file.
func queryFromEnv(envVar string) (string, string) {
	for _, name := range []string{envVar, envVar + "_FILE"} {
		value, ok := os.LookupEnv(name)
		if !ok || strings.TrimSpace(value) == "" {
			continue
		}
		if strings.HasSuffix(name, "_FILE") {
			content, err := os.ReadFile(value)
			if err != nil {
				panic(fmt.Sprintf("discovery filter %s points at %s, which cannot be read: %s", name, value, err))
			}
			return string(content), name
		}
		return value, name
	}
	return "", ""
}

func predicateString(predicate extquery.Predicate) string {
	if predicate == nil {
		return "<none>"
	}
	return predicate.String()
}

// logDropped reports how many records the filter removed. Without it a missing target is
// undebuggable — the extension simply never mentions it.
//
// Trace, not Debug: this runs on every discovery request, and only cached discoveries get an ETag
// short-circuit, so for the rest it repeats at the call interval forever with the same numbers.
// That matches CachedDiscovery, which traces the per-cycle line and debugs the outcome.
func (f *discoveryFilter) logDropped(before, after int, what string) {
	if dropped := before - after; dropped > 0 {
		log.Trace().
			Str("discoveryId", f.discoveryId).
			Msgf("Discovery filter dropped %d of %d %s.", dropped, before, what)
	}
}
