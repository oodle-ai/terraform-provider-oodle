package notifier

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type rootlyConfigModel struct {
	notifierConfigCommonModel
	BearerToken types.String `tfsdk:"bearer_token"`
	URL         types.String `tfsdk:"url"`
	// Payload holds a JSON object, same as the generic webhook notifier --
	// Rootly is a webhook underneath. See webhookConfigModel for why this is a
	// normalized JSON string rather than a typed map.
	Payload jsontypes.Normalized `tfsdk:"payload"`
}
