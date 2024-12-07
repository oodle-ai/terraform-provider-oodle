package notifier

import "github.com/hashicorp/terraform-plugin-framework/types"

type notifierConfigCommonModel struct {
	SendResolved types.Bool `tfsdk:"send_resolved"`
}
