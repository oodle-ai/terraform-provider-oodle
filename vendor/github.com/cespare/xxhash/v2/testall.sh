#!/bin/bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

set -eu -o pipefail

# Small convenience script for running the tests with various combinations of
# arch/tags. This assumes we're running on amd64 and have qemu available.

go test ./...
go test -tags purego ./...
GOARCH=arm64 go test
GOARCH=arm64 go test -tags purego
