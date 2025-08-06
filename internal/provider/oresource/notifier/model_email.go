package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type emailConfigModel struct {
	notifierConfigCommonModel
	To types.String `tfsdk:"to"`
}
