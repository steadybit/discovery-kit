// Package discovery_kit_api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.0 DO NOT EDIT.
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
	Regex      AttributeMatcher = "regex"
	StartsWith AttributeMatcher = "starts_with"
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
}

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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	merged, err := runtime.JSONMerge(t.union, b)
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

	"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",
	"qqX0IZPX/CGou/8EAAD//+gNxbCNKAAA",
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
