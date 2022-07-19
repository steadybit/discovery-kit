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