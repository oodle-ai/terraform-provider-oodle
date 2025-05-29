package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type pagerdutyConfigModel struct {
	notifierConfigCommonModel
	ServiceKey types.String `tfsdk:"service_key"`
	RoutingKey types.String `tfsdk:"routing_key"`
}
