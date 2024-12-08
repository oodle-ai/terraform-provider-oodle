// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssocreds

import "os"

func getHomeDirectory() string {
	return os.Getenv("USERPROFILE")
}
