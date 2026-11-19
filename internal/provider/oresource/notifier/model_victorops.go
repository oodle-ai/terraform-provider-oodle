package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type victorOpsConfigModel struct {
	notifierConfigCommonModel
	APIKey     types.String `tfsdk:"api_key"`
	RoutingKey types.String `tfsdk:"routing_key"`
}
