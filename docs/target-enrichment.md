# Target Enrichment

Targets carry attributes. Through these attributes, the targets are labeled and searchable. For example, to select a set of attack targets within an experiment. Target enrichment is the process of copying attributes from one target to another. This document describes the target enrichment process.

## Use Cases

In the following cases, you may choose to leverage target enrichments:

- Adding attributes to targets reported by extensions that you do not maintain.
- Adding attributes to targets reported by your extension
  - without duplicating the attribute-gathering logic
  - copying attributes from targets reported outside your extension

## Example

Let us start with a more detailed example before diving into the implementation details.

### Use Case

We assume we want to attack containers within a specific AWS availability zone. For example, we could stop or stress containers within zone `eu-central-1b`. However, containers have no native concept of availability zones! So how could we write a target selector for a zone? This is where target enrichments come in!

We do know that containers are running on EC2 instances and that those instances carry the attributes we desire. Suppose we copy the attributes from the EC2 instance target to (only those) containers running on that instance. In that case, we can leverage the AWS-specific attributes for container actions in experiments!

### Implementation

Now, how would we do this? The logic for the EC2 instance discovery resides in the AWS extension and not within our fictitious Docker container extension. Also, what about GCP, Azure et al., that we also want to support down the road? This is where target enrichment rules come in!

Target enrichment rules can be defined by as part of a Target enrichment rule set which can be added to [the root of your discovery](./discovery-api.md#target-enrichment-rules). Let us first see the full implementation, and then we will dissect it.

```yaml
{
  "id": "com.steadybit.extension_aws.aws.ec2-to-container",
  "version": "1.0.0",
  "src": {
    "type": "com.steadybit.extension_aws.aws.ec2-instance",
    "selector": {
      "aws-ec2.hostname": "${dest.container.host}"
    }
  },
  "dest": {
    "type": "com.steadybit.extension_container.container",
    "selector": {
      "container.host": "${src.aws-ec2.hostname}"
    }
  },
  "attributes": [
    {
      "matcher": "EQUALS",
      "name": "aws.account"
    },
    {
      "matcher": "EQUALS",
      "name": "aws.region"
    },
    {
      "matcher": "EQUALS",
      "name": "aws.zone"
    }
  ]
}
```

This example defines a rule that copies the attributes `aws.account`, `aws.region` and `aws.zone` from EC2 instances (`src.type`) to containers (`dest.type`).

Source and destination target selectors are required. These selectors facilitate matching between two targets. In our example, they express that a source EC2 instance's `aws-ec2.hostname` attribute must match a destination container's `container.host` attribute.

This concept is similar to [Kubernetes label selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).

Note that `$.enrichmentRules[*].src.type` and `$.enrichmentRules[*].dest.type` do not have to match `$.id`. This means that you could even implement multiple-consecutive copy operations.

### TargetEnrichmentRule Matcher
- EQUALS: The source attribute name must be equal to the attribute name provided.
- REGEX: The source attribute name must match the regular expression provided. You must use Postgres jsonb [regex syntax](https://www.postgresql.org/docs/current/functions-matching.html#POSIX-SYNTAX-DETAILS).
- CONTAINS: The source attribute name must contain the attribute name provided.
- STARTS_WITH: The source attribute name must start with the attribute name provided.

## FAQ

### I have changed target enrichment rules, but nothing happens in Steadybit?

When in doubt, ensure that the targets are re-reported by restarting the agents.
