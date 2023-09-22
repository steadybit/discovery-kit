# DiscoveryKit Go API

This module exposes Go types that you will find helpful when implementing an DiscoveryKit extension.

The types are generated automatically from the DiscoveryKit [OpenAPI specification](https://github.com/steadybit/discovery-kit/tree/main/openapi).

## Installation

Add the following to your `go.mod` file:

```
go get github.com/steadybit/discovery-kit/go/discovery_kit_api@v0.1.0
```

## Usage

```go
import (
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
)

DiscoveryList := discovery_kit_api.DiscoveryList{
    Discoverys: []discovery_kit_api.DescribingEndpointReference{
        {
            "GET",
            "/Discoverys/rollout-restart",
        },
    },
}
```

### Apply Deny List to Target Attributes

Use the function ApplyAttributeExcludes to filter out attributes from the targets that should not be exposed to the steadybit platform and should not be used the enhance other targets.

Excludes entries can be full attribute names and / or parts of the attribute key name with a trailing wildcard (*). (e.g.: ```aws.label.*```)

```go
import (
    "github.com/steadybit/discovery-kit/go/discovery_kit_api"
)
excludes := []string{"aws.label.*", "aws-ec2.instance.id"}] // From config or env variable

func getTargets(w http.ResponseWriter, _ *http.Request, _ []byte) {
      exthttp.WriteBody(w, discovery_kit_api.DiscoveryData{Targets: discovery_kit_api.ApplyAttributeExcludes(targets, excludes)})
}
```

