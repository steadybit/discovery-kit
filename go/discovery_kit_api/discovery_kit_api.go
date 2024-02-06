// Package discovery_kit_api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.2 DO NOT EDIT.
package discovery_kit_api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/runtime"
)

// Defines values for AttributeMatcher.
const (
	Contains   AttributeMatcher = "contains"
	Equals     AttributeMatcher = "equals"
	StartsWith AttributeMatcher = "starts_with"
)

// Defines values for DiscoveryDescriptionRestrictTo.
const (
	ANY    DiscoveryDescriptionRestrictTo = "ANY"
	LEADER DiscoveryDescriptionRestrictTo = "LEADER"
)

// Defines values for OrderByDirection.
const (
	ASC  OrderByDirection = "ASC"
	DESC OrderByDirection = "DESC"
)

// Defines values for ReadHttpMethod.
const (
	GET ReadHttpMethod = "GET"
)

// Attribute defines model for Attribute.
type Attribute struct {
	Matcher AttributeMatcher `json:"matcher"`
	Name    string           `json:"name"`
}

// AttributeMatcher defines model for Attribute.Matcher.
type AttributeMatcher string

// AttributeDescription defines model for AttributeDescription.
type AttributeDescription struct {
	// Attribute The attribute name, for example `cat.name`
	Attribute string      `json:"attribute"`
	Label     PluralLabel `json:"label"`
}

// AttributeDescriptions defines model for AttributeDescriptions.
type AttributeDescriptions struct {
	Attributes []AttributeDescription `json:"attributes"`
}

// Column defines model for Column.
type Column struct {
	// Attribute The attribute which should be displayed in the column.
	Attribute string `json:"attribute"`

	// FallbackAttributes If the given attribute is empty, the fallbackAttributes are used. The first non-empty attribute will be displayed.
	FallbackAttributes *[]string `json:"fallbackAttributes,omitempty"`
}

// DescribingEndpointReference HTTP endpoint which the Steadybit platform/agent could communicate with.
type DescribingEndpointReference struct {
	Method ReadHttpMethod `json:"method"`

	// Path Absolute path of the HTTP endpoint.
	Path string `json:"path"`
}

// DescribingEndpointReferenceWithCallInterval defines model for DescribingEndpointReferenceWithCallInterval.
type DescribingEndpointReferenceWithCallInterval struct {
	// CallInterval At what frequency should the state endpoint be called? Takes durations in the format of `100ms` or `10s`.
	CallInterval *string        `json:"callInterval,omitempty"`
	Method       ReadHttpMethod `json:"method"`

	// Path Absolute path of the HTTP endpoint.
	Path string `json:"path"`
}

// DiscoveredTargets Deprecated: use `DiscoveryData` instead. The results of a discovery call.
type DiscoveredTargets struct {
	Targets []Target `json:"targets"`
}

// DiscoveryData The results of a discovery call
type DiscoveryData struct {
	EnrichmentData *[]EnrichmentData `json:"enrichmentData,omitempty"`
	Targets        *[]Target         `json:"targets,omitempty"`
}

// DiscoveryDescription Provides details about a discovery, e.g., what endpoint needs to be called to discover targets.
type DiscoveryDescription struct {
	// Discover HTTP endpoint which the Steadybit platform/agent could communicate with.
	Discover DescribingEndpointReferenceWithCallInterval `json:"discover"`

	// Id A technical ID that is used to uniquely identify this type of discovery. You will typically want to use something like `org.example.discoveries.my-fancy-discovery`.
	Id string `json:"id"`

	// RestrictTo Deprecated: has not effect anymore. If the agent is deployed as a daemonset in Kubernetes, should the discovery only be called from the leader agent? This can be helpful to avoid duplicate targets for every running agent.
	// Deprecated:
	RestrictTo *DiscoveryDescriptionRestrictTo `json:"restrictTo,omitempty"`
}

// DiscoveryDescriptionRestrictTo Deprecated: has not effect anymore. If the agent is deployed as a daemonset in Kubernetes, should the discovery only be called from the leader agent? This can be helpful to avoid duplicate targets for every running agent.
type DiscoveryDescriptionRestrictTo string

// DiscoveryKitError RFC 7807 Problem Details for HTTP APIs compliant response body for error scenarios
type DiscoveryKitError struct {
	// Detail A human-readable explanation specific to this occurrence of the problem.
	Detail *string `json:"detail,omitempty"`

	// Instance A URI reference that identifies the specific occurrence of the problem.
	Instance *string `json:"instance,omitempty"`

	// Title A short, human-readable summary of the problem type.
	Title string `json:"title"`

	// Type A URI reference that identifies the problem type.
	Type *string `json:"type,omitempty"`
}

// DiscoveryList Lists all discoveries that the platform/agent could execute.
type DiscoveryList struct {
	Discoveries           []DescribingEndpointReference `json:"discoveries"`
	TargetAttributes      []DescribingEndpointReference `json:"targetAttributes"`
	TargetEnrichmentRules []DescribingEndpointReference `json:"targetEnrichmentRules"`
	TargetTypes           []DescribingEndpointReference `json:"targetTypes"`
}

// EnrichmentData A single discovered enrichment data
type EnrichmentData struct {
	// Attributes These attributes contains the actual data provided through the discovery.  These attributes are used to find matching targets and can be copied to a target.
	Attributes map[string][]string `json:"attributes"`

	// EnrichmentDataType The type of the enrichment data. Will be used to find matching targets to enrich data.
	EnrichmentDataType string `json:"enrichmentDataType"`

	// Id The id of the enrichment data, needs to be unique per enrichment data type.
	Id string `json:"id"`
}

// OrderBy defines model for OrderBy.
type OrderBy struct {
	Attribute string           `json:"attribute"`
	Direction OrderByDirection `json:"direction"`
}

// OrderByDirection defines model for OrderBy.Direction.
type OrderByDirection string

// PluralLabel defines model for PluralLabel.
type PluralLabel struct {
	One   string `json:"one"`
	Other string `json:"other"`
}

// ReadHttpMethod defines model for ReadHttpMethod.
type ReadHttpMethod string

// SourceOrDestination defines model for SourceOrDestination.
type SourceOrDestination struct {
	// Selector To identify a source or a destination, we employ a mechanism similar to Kubernetes label selectors. When this instance represents a source, you can use the placeholder `${src.attribute}` to refer to target attributes of the destination. Note that you can use the placeholders `${src.attribute}` and `${dest.attribute}` respectively.
	Selector map[string]string `json:"selector"`

	// Type The source or destination target type.
	Type string `json:"type"`
}

// Table defines model for Table.
type Table struct {
	Columns []Column  `json:"columns"`
	OrderBy []OrderBy `json:"orderBy"`
}

// Target A single discovered target
type Target struct {
	// Attributes These attributes include detailed information about the target provided through the discovery. These attributes are typically used as additional parameters within the attack implementation.
	Attributes map[string][]string `json:"attributes"`

	// Id The id of the target, needs to be unique per target type.
	Id string `json:"id"`

	// Label A label, which will be used by the platform to display the target
	Label string `json:"label"`

	// TargetType The type of the target. Will be used to find matching attacks and find the right ui configuration to show and select the targets.
	TargetType string `json:"targetType"`
}

// TargetDescription A definition of a target type and how it will be handled by the ui
type TargetDescription struct {
	// Category A human readable label categorizing the target type, e.g., 'cloud' or 'Kubernetes'.
	Category *string `json:"category,omitempty"`

	// Icon An icon that is used to identify the targets in the ui. Needs to be a data-uri containing an image.
	Icon *string `json:"icon,omitempty"`

	// Id a global unique name of the target type
	Id    string      `json:"id"`
	Label PluralLabel `json:"label"`
	Table Table       `json:"table"`

	// Version The version of the target type. Remember to increase the value everytime you update the definitions. The platform will ignore any definition changes with the same version. We do recommend usage of semver strings.
	Version string `json:"version"`
}

// TargetEnrichmentRule A rule describing how to enrich a target with data from another target or from enrichment data
type TargetEnrichmentRule struct {
	Attributes []Attribute         `json:"attributes"`
	Dest       SourceOrDestination `json:"dest"`

	// Id a global unique name of the enrichment rule
	Id  string              `json:"id"`
	Src SourceOrDestination `json:"src"`

	// Version The version of the enrichment rule. Remember to increase the value everytime you update the definitions. The platform will ignore any definition changes with the same version. We do recommend usage of semver strings.
	Version string `json:"version"`
}

// DescribeAttributesResponse defines model for DescribeAttributesResponse.
type DescribeAttributesResponse struct {
	union json.RawMessage
}

// DescribeTargetEnrichmentRulesResponse defines model for DescribeTargetEnrichmentRulesResponse.
type DescribeTargetEnrichmentRulesResponse struct {
	union json.RawMessage
}

// DescribeTargetResponse defines model for DescribeTargetResponse.
type DescribeTargetResponse struct {
	union json.RawMessage
}

// DiscoveryDescriptionResponse defines model for DiscoveryDescriptionResponse.
type DiscoveryDescriptionResponse struct {
	union json.RawMessage
}

// DiscoveryListResponse defines model for DiscoveryListResponse.
type DiscoveryListResponse struct {
	union json.RawMessage
}

// DiscoveryResponse defines model for DiscoveryResponse.
type DiscoveryResponse struct {
	union json.RawMessage
}

// AsAttributeDescriptions returns the union data inside the DescribeAttributesResponse as a AttributeDescriptions
func (t DescribeAttributesResponse) AsAttributeDescriptions() (AttributeDescriptions, error) {
	var body AttributeDescriptions
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromAttributeDescriptions overwrites any union data inside the DescribeAttributesResponse as the provided AttributeDescriptions
func (t *DescribeAttributesResponse) FromAttributeDescriptions(v AttributeDescriptions) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeAttributeDescriptions performs a merge with any union data inside the DescribeAttributesResponse, using the provided AttributeDescriptions
func (t *DescribeAttributesResponse) MergeAttributeDescriptions(v AttributeDescriptions) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

// AsDiscoveryKitError returns the union data inside the DescribeAttributesResponse as a DiscoveryKitError
func (t DescribeAttributesResponse) AsDiscoveryKitError() (DiscoveryKitError, error) {
	var body DiscoveryKitError
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDiscoveryKitError overwrites any union data inside the DescribeAttributesResponse as the provided DiscoveryKitError
func (t *DescribeAttributesResponse) FromDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDiscoveryKitError performs a merge with any union data inside the DescribeAttributesResponse, using the provided DiscoveryKitError
func (t *DescribeAttributesResponse) MergeDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

func (t DescribeAttributesResponse) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *DescribeAttributesResponse) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}

// AsTargetEnrichmentRule returns the union data inside the DescribeTargetEnrichmentRulesResponse as a TargetEnrichmentRule
func (t DescribeTargetEnrichmentRulesResponse) AsTargetEnrichmentRule() (TargetEnrichmentRule, error) {
	var body TargetEnrichmentRule
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromTargetEnrichmentRule overwrites any union data inside the DescribeTargetEnrichmentRulesResponse as the provided TargetEnrichmentRule
func (t *DescribeTargetEnrichmentRulesResponse) FromTargetEnrichmentRule(v TargetEnrichmentRule) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeTargetEnrichmentRule performs a merge with any union data inside the DescribeTargetEnrichmentRulesResponse, using the provided TargetEnrichmentRule
func (t *DescribeTargetEnrichmentRulesResponse) MergeTargetEnrichmentRule(v TargetEnrichmentRule) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

// AsDiscoveryKitError returns the union data inside the DescribeTargetEnrichmentRulesResponse as a DiscoveryKitError
func (t DescribeTargetEnrichmentRulesResponse) AsDiscoveryKitError() (DiscoveryKitError, error) {
	var body DiscoveryKitError
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDiscoveryKitError overwrites any union data inside the DescribeTargetEnrichmentRulesResponse as the provided DiscoveryKitError
func (t *DescribeTargetEnrichmentRulesResponse) FromDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDiscoveryKitError performs a merge with any union data inside the DescribeTargetEnrichmentRulesResponse, using the provided DiscoveryKitError
func (t *DescribeTargetEnrichmentRulesResponse) MergeDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

func (t DescribeTargetEnrichmentRulesResponse) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *DescribeTargetEnrichmentRulesResponse) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}

// AsTargetDescription returns the union data inside the DescribeTargetResponse as a TargetDescription
func (t DescribeTargetResponse) AsTargetDescription() (TargetDescription, error) {
	var body TargetDescription
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromTargetDescription overwrites any union data inside the DescribeTargetResponse as the provided TargetDescription
func (t *DescribeTargetResponse) FromTargetDescription(v TargetDescription) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeTargetDescription performs a merge with any union data inside the DescribeTargetResponse, using the provided TargetDescription
func (t *DescribeTargetResponse) MergeTargetDescription(v TargetDescription) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

// AsDiscoveryKitError returns the union data inside the DescribeTargetResponse as a DiscoveryKitError
func (t DescribeTargetResponse) AsDiscoveryKitError() (DiscoveryKitError, error) {
	var body DiscoveryKitError
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDiscoveryKitError overwrites any union data inside the DescribeTargetResponse as the provided DiscoveryKitError
func (t *DescribeTargetResponse) FromDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDiscoveryKitError performs a merge with any union data inside the DescribeTargetResponse, using the provided DiscoveryKitError
func (t *DescribeTargetResponse) MergeDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

func (t DescribeTargetResponse) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *DescribeTargetResponse) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}

// AsDiscoveryDescription returns the union data inside the DiscoveryDescriptionResponse as a DiscoveryDescription
func (t DiscoveryDescriptionResponse) AsDiscoveryDescription() (DiscoveryDescription, error) {
	var body DiscoveryDescription
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDiscoveryDescription overwrites any union data inside the DiscoveryDescriptionResponse as the provided DiscoveryDescription
func (t *DiscoveryDescriptionResponse) FromDiscoveryDescription(v DiscoveryDescription) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDiscoveryDescription performs a merge with any union data inside the DiscoveryDescriptionResponse, using the provided DiscoveryDescription
func (t *DiscoveryDescriptionResponse) MergeDiscoveryDescription(v DiscoveryDescription) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

// AsDiscoveryKitError returns the union data inside the DiscoveryDescriptionResponse as a DiscoveryKitError
func (t DiscoveryDescriptionResponse) AsDiscoveryKitError() (DiscoveryKitError, error) {
	var body DiscoveryKitError
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDiscoveryKitError overwrites any union data inside the DiscoveryDescriptionResponse as the provided DiscoveryKitError
func (t *DiscoveryDescriptionResponse) FromDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDiscoveryKitError performs a merge with any union data inside the DiscoveryDescriptionResponse, using the provided DiscoveryKitError
func (t *DiscoveryDescriptionResponse) MergeDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

func (t DiscoveryDescriptionResponse) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *DiscoveryDescriptionResponse) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}

// AsDiscoveryList returns the union data inside the DiscoveryListResponse as a DiscoveryList
func (t DiscoveryListResponse) AsDiscoveryList() (DiscoveryList, error) {
	var body DiscoveryList
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDiscoveryList overwrites any union data inside the DiscoveryListResponse as the provided DiscoveryList
func (t *DiscoveryListResponse) FromDiscoveryList(v DiscoveryList) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDiscoveryList performs a merge with any union data inside the DiscoveryListResponse, using the provided DiscoveryList
func (t *DiscoveryListResponse) MergeDiscoveryList(v DiscoveryList) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

// AsDiscoveryKitError returns the union data inside the DiscoveryListResponse as a DiscoveryKitError
func (t DiscoveryListResponse) AsDiscoveryKitError() (DiscoveryKitError, error) {
	var body DiscoveryKitError
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDiscoveryKitError overwrites any union data inside the DiscoveryListResponse as the provided DiscoveryKitError
func (t *DiscoveryListResponse) FromDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDiscoveryKitError performs a merge with any union data inside the DiscoveryListResponse, using the provided DiscoveryKitError
func (t *DiscoveryListResponse) MergeDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

func (t DiscoveryListResponse) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *DiscoveryListResponse) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}

// AsDiscoveryData returns the union data inside the DiscoveryResponse as a DiscoveryData
func (t DiscoveryResponse) AsDiscoveryData() (DiscoveryData, error) {
	var body DiscoveryData
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDiscoveryData overwrites any union data inside the DiscoveryResponse as the provided DiscoveryData
func (t *DiscoveryResponse) FromDiscoveryData(v DiscoveryData) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDiscoveryData performs a merge with any union data inside the DiscoveryResponse, using the provided DiscoveryData
func (t *DiscoveryResponse) MergeDiscoveryData(v DiscoveryData) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

// AsDiscoveryKitError returns the union data inside the DiscoveryResponse as a DiscoveryKitError
func (t DiscoveryResponse) AsDiscoveryKitError() (DiscoveryKitError, error) {
	var body DiscoveryKitError
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDiscoveryKitError overwrites any union data inside the DiscoveryResponse as the provided DiscoveryKitError
func (t *DiscoveryResponse) FromDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDiscoveryKitError performs a merge with any union data inside the DiscoveryResponse, using the provided DiscoveryKitError
func (t *DiscoveryResponse) MergeDiscoveryKitError(v DiscoveryKitError) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JsonMerge(t.union, b)
	t.union = merged
	return err
}

func (t DiscoveryResponse) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *DiscoveryResponse) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+RaW3PbuBX+K2fYzKQXRnamD9vRy44bu13PZjcex53Mzq5bQeCRiA0IcAHQDpvov3cO",
	"wDthSU7j9KF+sgTwXD6cy4dDfUy4LkqtUDmbLD8mBm2plUX/4RwtN2KNZ84Zsa4c2utmmVa5Vg6Vo39Z",
	"WUrBmRNanfxqtaLvLM+xYH5V1W82yfLnj8kzg5tkmfzupFd6EvbZk05J0FqSMJvs0v1PnQvL9R2a+nvh",
	"LozRJtnd7na7NMl6Kckyae2GjTbgcoSscQ1Y5xugykotlEvSxAknMVl2CEAPAXQY7NJu/YaZLboLZQTP",
	"C1TuupJPClZM35fC6iZHMGgr6SzoDTDATg2YSiJwJuUAo2AL9MYAWbMHpifHZRBBTwaKC14Pdj2Ay8CY",
	"MSat4sGGJ0Qmpu7psytrBexNrk5NHKDXwrqvgQzpeTpIOh+lsC4KR7eDDIlD8VUChDn2ZGnTwzDJllgM",
	"7NLGC9+NuhJMH0qjSzROhEZVMMdzNPQvqqpIlj8n+FvFpE1SDxMTiv61jhln/3UvXJ7cpomrS9JsnRFq",
	"SzArVnjhk4Vdmhj8rRIGM5Lsd6Wdzl6QXv+KnCIoifWyudVs6NAct24ZSGPqIwk/sKKUCCvO3IK+XiUR",
	"PyRboySh+07wSlaGydd+69TF3rJW2LFe2j1u+k/CYWEP2RbFb9eZwIxh9cNG26i1r7Ssiv/mFO5zwXOw",
	"ua5kBqG2lZLVmIFQPsW517CInciGSblm/P3ZCIuxtsuNl7IVd6gGaoUFLEpXp351LgiYQagsZgsgezfC",
	"WAdKqxf+qaH9QsqR3WRpdxwzk4/DOgp1U9iF2l40Ze4aN2hQ8QjI393cXHXlsAGZPH3rkGX1WjgoJXMb",
	"bYoTtiV2wf0BcF0UlaKiR565nJyZFAV0uc4Ohdo1suw758ofwu5dmpTM5XMzz9ZWS4KRlqmWkY0j270F",
	"zDk0tP+fJ4s/PpuHwgRHryttTb3ty6GX3KIHPXz70X0nXP6KSXmpHJo75msAk/KY6r/nyKgXjKHlEx0T",
	"qOgYmYMNeYqK123SEGTW0Yl1570OdBKzb+GGvUcLWWV8H7NtVtHJM0eIr16enhZ2BdrQv3Y1AfyXX7I/",
	"/V7ZT4X9ZD8Vn/JP2R/iBzCJ19v06SLymONsGh9mgTA2laE0SJKyZOlMhVMTz7v1JeU+rEbdewVCWTI2",
	"lIQ9DXieNa434qhSHYw+WDBasdFyMWIe0Sq8n0OMPeivK628oxy5GD82cyj9ktDswWDMGMZQXBl9JzLK",
	"EnRMSAtsrSs3hCQFXGwXaUjALoYVYmbB6T7d6EP7UHOXsfNgaHcc8vQxJWmXJiKLFA1wyHNKHwmX5+DI",
	"fmF9YyNbKyV+q1DWIDJUTmxqcLmwQDhSUHT+L+AnXYVW5+qSpMka7plyXohFsJqKrVBbkOI9wkqb7aJh",
	"VYtWikC7KOoXG6Z4/aITvYp2doP0P3c3+rF5mzMLSjvAzQa5A6bqQhtcQMMDQmkRdNal1EQzmKWTZlgQ",
	"O3ZUH7+v1mgUOrTpsMb2CaKVrAeHvjG68DsksgxN0PEt3BCUnCnamaMsN5UkuNidFhlkVbhbYBsmgYh6",
	"8aZSiqD0cgieln2f/fhTkiavL87OL64jVHtSHESWpH2s3cYuBcO02Jc+3XVkFmDXf3sF3/zl9Bu4Mnot",
	"sYDzJofIHV+fz64uLVXxUgoKmHYUBmud1cFpkgyWo2JGaDtPFy8xFtt5VTD1wiDL2Foi4IdSMuX7HNgS",
	"udgIToj7mNacV8bnTkszymBxNP6o0rMosTqDf1xfgmkTsUmpkD4CbejGrfLHKW3OZ67R5tq4dOqurYqC",
	"UTSOJPvkjYv3X3yOPwdET5uSd2MQb81NdjJ7mJtC31pgUsKgZASDvBUxeoAfkFcOHy6y4hEXpP10bdxv",
	"0iQUz8sgmMpS18zOHn83+1KqJ8PSr67/pi6/otZJ5A3PfGxQ5GQeQuz2odlRrEJezFjRLHeF2sq+eWA2",
	"HPxm9FC692bPskyQMCavRruOvWJGZkd2NJ9vJzmhPXJXMentoqwnWkTNz+hqm4+b4AJgJqq9M1PR3QiV",
	"gZ/mUCtrmxxTWdsSuS5F2NqOfRcxhMe88yZaw4jLtpyFjJwgvIB3zSV9v3FON0+Gp6J9IYtrF9kDutMR",
	"SwwRDCWa6b4ja6vv6RFI0kNzmjcmQ/PX+sCgZuZvJgzyljV3NOTtqyRNzi/evjpMQoYzr15YzMLh5Gxm",
	"pVZx+7RrBpT7raDH2823uzSZjCcGvv394iY6xHyrK8PxjTlH60TgGHMrLUrkLtCkh/J2Ejtvzt/E4myy",
	"S/cEnYH1ptB9nUHWm5PCPQIWxGqBQYE8Z0rYAqwohGSGIrAntuDnj9AabBfwLkcVmFLLfcAQrbZUsDut",
	"KdS68hlMrL/pyhxzLYn4rp59tIYvujPfrUir5xaeh4VXN4OC0eTMwIsF/Khdw0L2qLIxXVRcVs8+krTR",
	"90Q5KfDuUNbRGuMerCo91gMbW0eOpEMhP7vYiMX+DRG6eTyF0efx7bQZxkZ6gO7T/yhJbbk4NIBoLew1",
	"xN3zV/ajumPA9n/dFIXissqwmQj4YXSYmNHxh/EARWQTCIc6ZbRR9hdp35XoGtp5BSUzrEBHgX4v6GId",
	"urNzjL8HQddqKv8hY2IRfbhPBdMfbE/7Q3zwMmR6on4hbeZ798O+u65HNL6ZlZSS1QODoheXjsgdbv0N",
	"kzjQ8QOQgY74FXrUiG3uoBLEiDZi20xM6WGb63u/NyTxQJE9smUHuEauHGzZ87fuEbgz3Ajlo2b0At1D",
	"QhaT5cJ1B5Ezlcn+LCoxSzTOHG61qR+8b0N3AQ09pHlA/NtTqT4pyIR2bvacS11lz6mOPu+b0PM4xeJR",
	"TxXQwmyCNRhc9aOUJl0qsYAfB+HNPNN6URnRkl4fCwpEwbY+zAv24TWqrcuT5ctT/zeahdPjy+gLiHjG",
	"MdhKvWayTSzFikmgQtMdvsDLRgqupovsH5+u/Y9akjs0NhpUlFLNYsTYBVxjgcU6tHShuEHW9Oc7JisM",
	"MywnCvT9uyozP+TyXb4NVRsm6F0l8NEptkobCtp6GNTEYrYYqmCYrRCGjXkLeIeQEcPguihQZVBZtvUQ",
	"Wyzu0EAA9NgsbSEZ5uta4p7snPxWKJIz/rc9WXfP9QnZXzS6hPX++YuAnyYy5clqu6pN+PpxV8fHvRQ+",
	"8ppPPOiQzBhZ/owUmfxCKpYm1vDPNOYx4T8x5P8jBQja5rgPtCoSxLWyVUFx9/P8dzO3aUL8ibCWgmPz",
	"S5tuyGiSZfLD5U3S/lTEf+jnoUn/hnA4kIazq8uBwcvk5eJ0cerJbomKlSJZJn9evFychveZuU2WqpLS",
	"p0xW8YdM3f0nAAD//0XFpTK8KQAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
