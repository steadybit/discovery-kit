package validate

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/steadybit/discovery-kit/go/discovery_kit_test/client"
)

func ValidateEndpointReferences(path string, restyClient *resty.Client) error {
	c := client.NewDiscoveryClient(path, restyClient)
	var allErr []error

	list, err := c.ListDiscoveries()
	if err != nil {
		allErr = append(allErr, err)
	}

	for _, discovery := range list.Discoveries {
		description, err := c.DescribeDiscovery(discovery)
		if err == nil {
			_, err = c.Discover(description.Discover)
			if err != nil {
				allErr = append(allErr, err)
			}
		} else {
			allErr = append(allErr, err)
		}
	}

	for _, rule := range list.TargetEnrichmentRules {
		_, err := c.DescribeEnrichmentRule(rule)
		if err != nil {
			allErr = append(allErr, err)
		}
	}

	for _, target := range list.TargetTypes {
		_, err := c.DescribeTarget(target)
		if err != nil {
			allErr = append(allErr, err)
		}
	}

	for _, attribute := range list.TargetAttributes {
		_, err := c.DescribeAttributes(attribute)
		if err != nil {
			allErr = append(allErr, err)
		}
	}

	for _, rules := range list.TargetEnrichmentRules {
		_, err := c.DescribeEnrichmentRule(rules)
		if err != nil {
			allErr = append(allErr, err)
		}
	}

	return errors.Join(allErr...)
}
