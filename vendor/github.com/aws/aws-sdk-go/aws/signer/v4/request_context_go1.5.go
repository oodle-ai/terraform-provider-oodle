// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !go1.7
// +build !go1.7

package v4

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
)

func requestContext(r *http.Request) aws.Context {
	return aws.BackgroundContext()
}
