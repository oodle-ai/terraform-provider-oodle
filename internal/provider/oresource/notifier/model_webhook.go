package notifier

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type webhookConfigModel struct {
	notifierConfigCommonModel
	URL types.String `tfsdk:"url"`
	// Payload holds a JSON object. jsontypes.Normalized compares it
	// semantically, so re-encoding it on read does not diff against the
	// formatting used in the configuration.
	Payload jsontypes.Normalized `tfsdk:"payload"`
}
