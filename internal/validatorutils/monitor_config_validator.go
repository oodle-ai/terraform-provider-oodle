package validatorutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type monitorConfigValidator struct{}

var _ resource.ConfigValidator = (*monitorConfigValidator)(nil)

func NewMonitorConfigValidator() resource.ConfigValidator {
	return &monitorConfigValidator{}
}

func (m monitorConfigValidator) Description(ctx context.Context) string {
	return "Validates that only one notification configuration approach is used: either (notification_policy_id/label_matcher_notification_policies) or notifications"
}

func (m monitorConfigValidator) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m monitorConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	// Get the configuration values
	var notificationPolicyID types.String
	var labelMatcherNotificationPolicies types.List
	var notifications types.List

	// Extract values from config using proper paths
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx,
		path.Root("notification_policy_id"),
		&notificationPolicyID)...)

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx,
		path.Root("label_matcher_notification_policies"),
		&labelMatcherNotificationPolicies)...)

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx,
		path.Root("notifications"),
		&notifications)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if old approach is used
	oldApproachUsed := (!notificationPolicyID.IsNull() && !notificationPolicyID.IsUnknown() && notificationPolicyID.ValueString() != "") ||
		(!labelMatcherNotificationPolicies.IsNull() && !labelMatcherNotificationPolicies.IsUnknown() && len(labelMatcherNotificationPolicies.Elements()) > 0)

	// Check if new approach is used
	newApproachUsed := !notifications.IsNull() && !notifications.IsUnknown() && len(notifications.Elements()) > 0

	// Both approaches cannot be used simultaneously
	if oldApproachUsed && newApproachUsed {
		resp.Diagnostics.AddError(
			"Conflicting notification configurations",
			"only one of notification_policy_id / label_matcher_notification_policies (deprecated)or notifications (newer approach) should be set",
		)
	}
}
