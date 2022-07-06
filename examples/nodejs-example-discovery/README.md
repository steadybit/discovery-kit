# Discovery Example: Github-Gist

A tiny discovery implementation which exposes the required HTTP APIs and discovers some targets of type `cat`. If you are new to Steadybit's discoveries, this
example app will help you understand the fundamental contracts and control flows.

## Starting the example

```sh
npm install
npm start
```

## Starting the example using Docker

```sh
docker run -it \
  --rm \
  --init \
  -p 3002:3002 \
  ghcr.io/steadybit/example-nodejs-discovery:main
```