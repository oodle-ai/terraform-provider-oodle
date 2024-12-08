// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build go1.12
// +build go1.12

package shareddefaults

import (
	"os"
)

func userHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
