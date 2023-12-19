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
        "types": ["ACTION","DISCOVERY","EVENT"],                                                                                                                                                                           
        "protocol": "http"                                                                                                                                                                                               
      }                                                                                                                                                                                                          
    ]                                                                                                                                                                                                    
  }
```

## With Environment Variables

If you can't use the automatic annotation discovery, for example if you are not deploying to kubernetes, you can still register extensions using
environment variables. You can be specify them via `agent.env` files or directly via the command line.

Please note that these environment variables are index-based (referred to as `n`) to register multiple extension instances.

| Environment Variable<br/>(`n` refers to the index of the extension's instance) | Required | Description                                                                                                                                                                      |
|--------------------------------------------------------------------------------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_n_URL`                                 | yes      | Fully-qualified URL of the endpoint to get the [index response](./discovery-api.md#index-response) of an extension, e.g., `http://discoveries.steadybit.svc.cluster.local:8080/` |
| `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_n_METHOD`                              |          | Optional HTTP method to use. Default: `GET`                                                                                                                                      |
| `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_n_BASIC_USERNAME`                      |          | Optional basic authentication username to use within HTTP requests.                                                                                                              |
| `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_n_BASIC_PASSWORD`                      |          | Optional basic authentication password to use within HTTP requests.                                                                                                              |

### Example
To register, e.g., two extensions, where the second one requires basic authentication, you use
- `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_URL`,
- `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_1_URL`,
- `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_1_BASIC_USERNAME` and
- `STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_1_BASIC_USERNAME`.

