// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import "github.com/hashicorp/terraform-plugin-framework/types"

type grouping struct {
	ByMonitor types.Bool `tfsdk:"by_monitor"`
	ByLabels  types.List `tfsdk:"by_labels"`
	Disabled  types.Bool `tfsdk:"disabled"`
}
