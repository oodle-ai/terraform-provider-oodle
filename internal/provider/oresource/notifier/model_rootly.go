package notifier

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/resourceutils"
)

type rootlyConfigModel struct {
	notifierConfigCommonModel
	BearerToken types.String `tfsdk:"bearer_token"`
	URL         types.String `tfsdk:"url"`
	// Payload holds a JSON object. jsontypes.Normalized compares it
	// semantically, so re-encoding it on read does not diff against the
	// formatting used in the configuration.
	Payload resourceutils.JSON `tfsdk:"payload"`
}
