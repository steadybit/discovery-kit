# Extension Registration

Steadybit's agents need to be told where they can find extensions.

## With Automatic Kubernetes Annotation Discovery

The annotation discovery mechanism is based on the following annotations on service or daemonset level:

``` 
steadybit.com/extension-auto-discovery:                                                                                                                                                                              
  {                                                                                                                                                                                                                
    "extensions": [                                                                                                                                                                                                
      {                                                                                                                                                                                                            
        "port": 8088,                                                                                                                                                                                              
        "types": ["ACTION","DISCOVERY","EVENTS"],                                                                                                                                                                           
        "protocol": "http"                                                                                                                                                                                               
      }                                                                                                                                                                                                          
    ]                                                                                                                                                                                                    
  }
```

If you are using our helm charts, the annotations are automatically added to the service or daemonset definitions of the extension.

## With Environment Variables

If you can't use the automatic annotation discovery, for example if you are not deploying to kubernetes, you can still register extensions using environment
variables.

This can be done via `agent.env` files or directly via the command line.

The environment variables are:

- `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_URL`: Required fully-qualified URL defining which HTTP URL should be requested to get
  the [index response](./discovery-api.md#index-response), e.g., `http://discoveries.steadybit.svc.cluster.local:8080/`.
- `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_METHOD`: Optional HTTP method to use. Defaults to GET.
- `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_BASIC_USERNAME`: Optional basic authentication username to use within HTTP requests.
- `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_BASIC_PASSWORD`: Optional basic authentication password to use within HTTP requests.

These environment variables can occur multiple times with different indices to register multiple discovery providers,
e.g., `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_URL` and `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_1_URL`.
