// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build go1.7
// +build go1.7

package credentials

import "context"

// backgroundContext returns a context that will never be canceled, has no
// values, and no deadline. This context is used by the SDK to provide
// backwards compatibility with non-context API operations and functionality.
//
// Go 1.6 and before:
// This context function is equivalent to context.Background in the Go stdlib.
//
// Go 1.7 and later:
// The context returned will be the value returned by context.Background()
//
// See https://golang.org/pkg/context for more information on Contexts.
func backgroundContext() Context {
	return context.Background()
}
