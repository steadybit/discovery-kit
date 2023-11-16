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

	// RestrictTo If the agent is deployed as a daemonset in Kubernetes, should the discovery only be called from the leader agent? This can be helpful to avoid duplicate targets for every running agent.
	RestrictTo *DiscoveryDescriptionRestrictTo `json:"restrictTo,omitempty"`
}

// DiscoveryDescriptionRestrictTo If the agent is deployed as a daemonset in Kubernetes, should the discovery only be called from the leader agent? This can be helpful to avoid duplicate targets for every running agent.
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

	"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",
	"KQAA",
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
