package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type msTeamsV2ConfigModel struct {
	notifierConfigCommonModel
	WebhookURL types.String `tfsdk:"webhook_url"`
	Title      types.String `tfsdk:"title"`
	Text       types.String `tfsdk:"text"`
}
