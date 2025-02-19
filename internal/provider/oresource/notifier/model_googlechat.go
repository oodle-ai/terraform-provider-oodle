package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type googleChatConfigModel struct {
	notifierConfigCommonModel
	URL       types.String `tfsdk:"url"`
	Threading types.Bool   `tfsdk:"threading"`
}
