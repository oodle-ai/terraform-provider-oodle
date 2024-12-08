// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !windows
// +build !windows

package ssocreds

import "os"

func getHomeDirectory() string {
	return os.Getenv("HOME")
}
