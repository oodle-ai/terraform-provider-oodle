package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type slackConfigModel struct {
	notifierConfigCommonModel
	APIURL    types.String `tfsdk:"api_url"`
	Channel   types.String `tfsdk:"channel"`
	TitleLink types.String `tfsdk:"title_link"`
	Text      types.String `tfsdk:"text"`
}
