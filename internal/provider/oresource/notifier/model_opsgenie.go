// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type opsgenieConfigModel struct {
	notifierConfigCommonModel
	APIKey types.String `tfsdk:"api_key"`
}
