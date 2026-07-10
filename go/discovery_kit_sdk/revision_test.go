// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package discovery_kit_sdk

import (
	"testing"

	"github.com/steadybit/extension-kit/exthttp"
	"github.com/stretchr/testify/assert"
)

func TestRegisterBumpsRevision(t *testing.T) {
	ClearRegisteredDiscoveries()
	t.Cleanup(ClearRegisteredDiscoveries)

	before := exthttp.Revision()
	Register(newMockTargetDiscovery())
	assert.NotEqual(t, before, exthttp.Revision(), "Register must bump the index revision")
}

func TestClearRegisteredDiscoveriesBumpsRevision(t *testing.T) {
	before := exthttp.Revision()
	ClearRegisteredDiscoveries()
	assert.NotEqual(t, before, exthttp.Revision(), "ClearRegisteredDiscoveries must bump the index revision")
}
