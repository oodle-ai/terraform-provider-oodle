// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type pagerdutyConfigModel struct {
	notifierConfigCommonModel
	ServiceKey types.String `tfsdk:"service_key"`
}
