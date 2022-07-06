# Discovery Example: Github-Gist

A tiny discovery implementation which exposes the required HTTP APIs and discovers some targets of type `cat`. If you are new to Steadybit's discoveries, this
example app will help you understand the fundamental contracts and control flows.

## Starting the example

```sh
npm install
npm start
```

## Starting the example through Kubernetes

This is the recommended approach to give this example app a try. The app is deployed within a namespace `example-nodejs-pet-discovery` as a deployment
called `example-nodejs-pet-discovery`. For more details, please inspect `kubernetes.yml`.

```sh
kubectl apply -f kubernetes.yml
```

Once deployed in your Kubernetes cluster the example is reachable
through `http://example-nodejs-pet-discovery.example-nodejs-pet-discovery.svc.cluster.local:8085`. Steadybit agents can be configured to support this
discovery provider through the environment variable `STEADYBIT_AGENT_ATTACKS_DISCOVERIES_0_URL`.

## Starting the example using Docker

```sh
docker run -it \
  --rm \
  --init \
  -p 8085:8085 \
  ghcr.io/steadybit/example-nodejs-pet-discovery:main
```