package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type rootlyConfigModel struct {
	notifierConfigCommonModel
	BearerToken types.String `tfsdk:"bearer_token"`
	URL         types.String `tfsdk:"url"`
}
