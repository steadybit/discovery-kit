#!/usr/bin/env bash

#
# Copyright 2024 steadybit GmbH. All rights reserved.
#

set -eo pipefail

go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.3.0
oapi-codegen -config generator-config.yml -o discovery_kit_api.go ../../openapi/spec.yml