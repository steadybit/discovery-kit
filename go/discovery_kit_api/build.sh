#!/usr/bin/env bash

set -eo pipefail

go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.16.2
oapi-codegen -config generator-config.yml ../../openapi/spec.yml > discovery_kit_api.go