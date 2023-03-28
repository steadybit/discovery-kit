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

Target enrichment rules can be defined as part of the target type description. Let us first see the full implementation, and then we will dissect it.

```yaml
{
  "id": "container",
  "version": "1.0.0",
  "label": {
    "one": "Container",
    "other": "Containers"
  },
  "icon": "data:image/svg+xml;base64,...",
  "table": {
    # ...
  },
  "enrichmentRules": [{
    "src": {
      "type": "ec2-instance",
      "selector": {
        "aws-ec2.hostname": "${dest.container.host}"
      }
    },
    "dest": {
      "type": "container",
      "selector": {
        "container.host": "${src.aws-ec2.hostname}"
      }
    },
    "attributes": [
      {
        "matcher": "equals",
        "name": "aws.account"
      },
      {
        "matcher": "equals",
        "name": "aws.region"
      },
      {
        "matcher": "equals",
        "name": "aws.zone"
      }
    ]
  }]
}
```

You can specify multiple target enrichment rules as part of every target type description. This example defines a rule that copies the attributes `aws.account`, `aws.region` and `aws.zone` from EC2 instances (`src.type`) to containers (`dest.type`).

Source and destination target selectors are required. These selectors facilitate matching between two targets. In our example, they express that a source EC2 instance's `aws-ec2.hostname` attribute must match a destination container's `container.host` attribute.

This concept is similar to [Kubernetes label selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).

Note that `$.enrichmentRules[*].src.type` and `$.enrichmentRules[*].dest.type` do not have to match `$.id`. This means that you could even implement multiple-consecutive copy operations.

## FAQ

### I have changed target enrichment rules, but nothing happens in Steadybit?

When in doubt, ensure that the targets are re-reported. You can either restart the agents or through the *re-trigger target discovery* button in Steadybit's UI (under Setting -> Agents).
