package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-resty/resty/v2"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type DiscoveryAPI interface {
	DiscoverTargets(discoveryId string) ([]discovery_kit_api.Target, error)
	DiscoverEnrichmentData(discoveryId string) ([]discovery_kit_api.EnrichmentData, error)

	ListDiscoveries() (discovery_kit_api.DiscoveryList, error)
	Discover(ref discovery_kit_api.DescribingEndpointReferenceWithCallInterval) (discovery_kit_api.DiscoveryData, error)
	DescribeDiscovery(ref discovery_kit_api.DescribingEndpointReference) (discovery_kit_api.DiscoveryDescription, error)
	DescribeTarget(ref discovery_kit_api.DescribingEndpointReference) (discovery_kit_api.TargetDescription, error)
	DescribeAttributes(ref discovery_kit_api.DescribingEndpointReference) (discovery_kit_api.AttributeDescriptions, error)
	DescribeEnrichmentRule(ref discovery_kit_api.DescribingEndpointReference) (discovery_kit_api.TargetEnrichmentRule, error)
}

type clientImpl struct {
	client   *resty.Client
	rootPath string
	spec     *openapi3.T
}

func NewDiscoveryClient(rootPath string, client *resty.Client) DiscoveryAPI {
	spec, _ := discovery_kit_api.GetSwagger()
	return &clientImpl{
		client:   client,
		rootPath: rootPath,
		spec:     spec,
	}
}

func (c *clientImpl) DiscoverTargets(discoveryId string) ([]discovery_kit_api.Target, error) {
	if d, err := c.discoverById(discoveryId); err != nil {
		return nil, err
	} else {
		return *d.Targets, nil
	}
}

func (c *clientImpl) DiscoverEnrichmentData(discoveryId string) ([]discovery_kit_api.EnrichmentData, error) {
	if d, err := c.discoverById(discoveryId); err != nil {
		return nil, err
	} else {
		return *d.EnrichmentData, nil
	}
}

func (c *clientImpl) discoverById(discoveryId string) (discovery_kit_api.DiscoveryData, error) {
	var data discovery_kit_api.DiscoveryData
	discoveries, err := c.ListDiscoveries()
	if err != nil {
		return data, err
	}

	for _, discovery := range discoveries.Discoveries {
		description, err := c.DescribeDiscovery(discovery)
		if err != nil {
			return data, err
		}

		if description.Id == discoveryId {
			data, err = c.Discover(description.Discover)
			return data, err
		}
	}

	return data, fmt.Errorf("discovery with id %s not found", discoveryId)
}

func (c *clientImpl) Discover(ref discovery_kit_api.DescribingEndpointReferenceWithCallInterval) (discovery_kit_api.DiscoveryData, error) {
	var data discovery_kit_api.DiscoveryData
	err := c.executeAndValidate(
		discovery_kit_api.DescribingEndpointReference{Method: discovery_kit_api.DescribingEndpointReferenceMethod(ref.Method), Path: ref.Path},
		&data,
		"DiscoveryData",
	)
	return data, err
}

func (c *clientImpl) ListDiscoveries() (discovery_kit_api.DiscoveryList, error) {
	var list discovery_kit_api.DiscoveryList
	err := c.executeAndValidate(discovery_kit_api.DescribingEndpointReference{Path: c.rootPath}, &list, "DiscoveryList")
	return list, err
}

func (c *clientImpl) DescribeDiscovery(ref discovery_kit_api.DescribingEndpointReference) (discovery_kit_api.DiscoveryDescription, error) {
	var description discovery_kit_api.DiscoveryDescription
	err := c.executeAndValidate(ref, &description, "DiscoveryDescription")
	return description, err
}

func (c *clientImpl) DescribeTarget(ref discovery_kit_api.DescribingEndpointReference) (discovery_kit_api.TargetDescription, error) {
	var description discovery_kit_api.TargetDescription
	err := c.executeAndValidate(ref, &description, "TargetDescription")
	return description, err
}

func (c *clientImpl) DescribeAttributes(ref discovery_kit_api.DescribingEndpointReference) (discovery_kit_api.AttributeDescriptions, error) {
	var descriptions discovery_kit_api.AttributeDescriptions
	err := c.executeAndValidate(ref, &descriptions, "AttributeDescriptions")
	return descriptions, err
}

func (c *clientImpl) DescribeEnrichmentRule(ref discovery_kit_api.DescribingEndpointReference) (discovery_kit_api.TargetEnrichmentRule, error) {
	var enrichmentRule discovery_kit_api.TargetEnrichmentRule
	err := c.executeAndValidate(ref, &enrichmentRule, "TargetEnrichmentRule")
	return enrichmentRule, err
}

func (c *clientImpl) executeAndValidate(ref discovery_kit_api.DescribingEndpointReference, result interface{}, schemaName string) error {
	method, path := getMethodAndPath(ref)
	res, err := c.client.R().SetResult(result).Execute(method, path)
	if err != nil {
		return fmt.Errorf("%s %s failed: %w", method, path, err)
	}
	if !res.IsSuccess() {
		return fmt.Errorf("%s %s failed: %d %s", method, path, res.StatusCode(), res.Body())
	}

	return c.validateResponseBody(schemaName, res.Body())
}

func (c *clientImpl) validateResponseBody(name string, body []byte) error {
	if c.spec == nil || name == "" {
		return nil
	}

	schema, ok := c.spec.Components.Schemas[name]
	if !ok {
		return fmt.Errorf("component schema '%s' not found", name)
	}

	var decoded interface{}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	err := dec.Decode(&decoded)
	if err != nil {
		return fmt.Errorf("error decoding response body: %w", err)
	}

	err = schema.Value.VisitJSON(decoded, openapi3.VisitAsResponse())
	if err != nil {
		return fmt.Errorf("error validating response body using schema '%s': %w", name, err)
	}
	return nil
}

func getMethodAndPath(ref discovery_kit_api.DescribingEndpointReference) (string, string) {
	method := "GET"
	if len(ref.Method) > 0 {
		method = cases.Upper(language.English).String(string(ref.Method))
	}
	return method, ref.Path
}
