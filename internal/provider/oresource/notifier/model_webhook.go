package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type webhookConfigModel struct {
	notifierConfigCommonModel
	URL types.String `tfsdk:"url"`
}
