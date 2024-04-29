#!/usr/bin/env bash

set -eo pipefail

go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.1.0
oapi-codegen -config generator-config.yml ../../openapi/spec.yml > discovery_kit_api.go