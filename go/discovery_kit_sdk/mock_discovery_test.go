// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package discovery_kit_sdk

import (
	"context"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/extutil"
	"sync"
	"time"
)
import "github.com/stretchr/testify/mock"

type MockDiscovery struct {
	mock.Mock
	Now  func() time.Time
	cond *sync.Cond
}

type MockTargetDiscovery struct {
	MockDiscovery
}

var (
	_ TargetDiscovery    = (*MockTargetDiscovery)(nil)
	_ TargetDescriber    = (*MockTargetDiscovery)(nil)
	_ AttributeDescriber = (*MockTargetDiscovery)(nil)
)

func (e *MockDiscovery) WaitForNextDiscovery(fn ...func()) {
	e.cond.L.Lock()
	defer e.cond.L.Unlock()
	go func() {
		for _, f := range fn {
			f()
		}
	}()
	e.cond.Wait()
}

func (e *MockDiscovery) Describe() discovery_kit_api.DiscoveryDescription {
	args := e.Called()
	return args.Get(0).(discovery_kit_api.DiscoveryDescription)
}

func (e *MockTargetDiscovery) DiscoverTargets(ctx context.Context) []discovery_kit_api.Target {
	args := e.Called(ctx)
	e.cond.L.Lock()
	defer e.cond.L.Unlock()
	e.cond.Broadcast()
	return args.Get(0).([]discovery_kit_api.Target)
}

func (e *MockTargetDiscovery) DescribeTarget() discovery_kit_api.TargetDescription {
	args := e.Called()
	return args.Get(0).(discovery_kit_api.TargetDescription)
}

func (e *MockTargetDiscovery) DescribeAttributes() []discovery_kit_api.AttributeDescription {
	args := e.Called()
	return args.Get(0).([]discovery_kit_api.AttributeDescription)
}

func newMockTargetDiscovery() *MockTargetDiscovery {
	m := &MockTargetDiscovery{MockDiscovery{Now: time.Now, cond: sync.NewCond(&sync.Mutex{})}}
	m.On("Describe").Return(discovery_kit_api.DiscoveryDescription{
		Id: "example",
	})
	m.On("DescribeTarget").Return(discovery_kit_api.TargetDescription{
		Category: extutil.Ptr("examples"),
		Id:       "example",
		Label:    discovery_kit_api.PluralLabel{One: "Example Target", Other: "Example Targets"},
		Version:  "unknown",
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{},
			OrderBy: []discovery_kit_api.OrderBy{},
		},
	})
	m.On("DescribeAttributes").Return([]discovery_kit_api.AttributeDescription{
		{
			Attribute: "target.created",
			Label: discovery_kit_api.PluralLabel{
				One:   "Creation Date",
				Other: "Creation Dates",
			},
		},
	})
	call := m.On("DiscoverTargets", mock.Anything)
	call.RunFn = func(args mock.Arguments) {
		call.ReturnArguments = mock.Arguments{[]discovery_kit_api.Target{
			{
				Id:         "target",
				TargetType: "example",
				Label:      "Example Target",
				Attributes: map[string][]string{
					"example.created": {m.Now().String()},
				},
			},
		}}
	}
	return m
}

type MockEnrichmentDataDiscovery struct {
	MockDiscovery
}

var (
	_ EnrichmentDataDiscovery  = (*MockEnrichmentDataDiscovery)(nil)
	_ EnrichmentRulesDescriber = (*MockEnrichmentDataDiscovery)(nil)
	_ AttributeDescriber       = (*MockEnrichmentDataDiscovery)(nil)
)

func (e *MockEnrichmentDataDiscovery) DiscoverEnrichmentData(ctx context.Context) []discovery_kit_api.EnrichmentData {
	args := e.Called(ctx)
	e.cond.L.Lock()
	defer e.cond.L.Unlock()
	e.cond.Broadcast()
	return args.Get(0).([]discovery_kit_api.EnrichmentData)
}

func (e *MockEnrichmentDataDiscovery) DescribeAttributes() []discovery_kit_api.AttributeDescription {
	args := e.Called()
	return args.Get(0).([]discovery_kit_api.AttributeDescription)
}

func (e *MockEnrichmentDataDiscovery) DescribeEnrichmentRule() discovery_kit_api.TargetEnrichmentRule {
	args := e.Called()
	return args.Get(0).(discovery_kit_api.TargetEnrichmentRule)
}

func newMockEnrichmentDataDiscovery() *MockEnrichmentDataDiscovery {
	m := &MockEnrichmentDataDiscovery{MockDiscovery{Now: time.Now, cond: sync.NewCond(&sync.Mutex{})}}
	m.On("Describe").Return(discovery_kit_api.DiscoveryDescription{
		Id: "example-ed",
	})
	m.On("DescribeEnrichmentRule").Return(discovery_kit_api.TargetEnrichmentRule{
		Src: discovery_kit_api.SourceOrDestination{
			Selector: map[string]string{},
			Type:     "example-ed",
		},
		Dest: discovery_kit_api.SourceOrDestination{
			Selector: map[string]string{},
			Type:     "other",
		},
		Id:         "enrichmentRule",
		Version:    "ed",
		Attributes: []discovery_kit_api.Attribute{},
	})
	m.On("DescribeAttributes").Return([]discovery_kit_api.AttributeDescription{
		{
			Attribute: "example-ed.created",
			Label: discovery_kit_api.PluralLabel{
				One:   "Creation Date",
				Other: "Creation Dates",
			},
		},
	})
	call := m.On("DiscoverEnrichmentData", mock.Anything)
	call.RunFn = func(args mock.Arguments) {
		call.ReturnArguments = mock.Arguments{[]discovery_kit_api.EnrichmentData{
			{
				Id:                 "example-ed",
				EnrichmentDataType: "example-ed",
				Attributes: map[string][]string{
					"example-ed.created": {m.Now().String()},
				},
			},
		}}
	}

	return m
}
