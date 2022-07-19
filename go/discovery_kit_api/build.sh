#!/usr/bin/env bash

set -eo pipefail

oapi-codegen -config generator-config.yml ../../openapi/spec.yml > discovery_kit_api.go

cat extras.go.txt >> discovery_kit_api.go