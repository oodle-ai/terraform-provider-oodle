package validatorutils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// syntheticMonitorConfigValidator validates, at plan time, that exactly one
// rule_config block is set and that it agrees with rule_type. Without this the
// mismatch is only caught by the server on apply, producing a worse UX.
type syntheticMonitorConfigValidator struct{}

var _ resource.ConfigValidator = (*syntheticMonitorConfigValidator)(nil)

// ruleTypeToConfigAttr maps each rule_type to the rule_config attribute that
// must be set for it.
var ruleTypeToConfigAttr = map[string]string{
	"http":      "http",
	"multistep": "multistep",
}

func NewSyntheticMonitorConfigValidator() resource.ConfigValidator {
	return &syntheticMonitorConfigValidator{}
}

func (v syntheticMonitorConfigValidator) Description(ctx context.Context) string {
	return "Validates that exactly one rule_config block is set and that it matches rule_type."
}

func (v syntheticMonitorConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v syntheticMonitorConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var ruleType types.String
	var httpConfig types.Object
	var multistepConfig types.Object

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx,
		path.Root("rule_type"), &ruleType)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx,
		path.Root("rule_config").AtName("http"), &httpConfig)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx,
		path.Root("rule_config").AtName("multistep"), &multistepConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpSet := !httpConfig.IsNull() && !httpConfig.IsUnknown()
	multistepSet := !multistepConfig.IsNull() && !multistepConfig.IsUnknown()

	// Exactly one rule_config block must be set.
	switch {
	case !httpSet && !multistepSet:
		resp.Diagnostics.AddAttributeError(
			path.Root("rule_config"),
			"Missing rule configuration",
			"Exactly one of rule_config.http or rule_config.multistep must be set.",
		)
		return
	case httpSet && multistepSet:
		resp.Diagnostics.AddAttributeError(
			path.Root("rule_config"),
			"Conflicting rule configuration",
			"Only one of rule_config.http or rule_config.multistep may be set.",
		)
		return
	}

	// When rule_type is known, the block that is set must match it. Unknown
	// rule_types (e.g. interpolated from another resource) are left to the
	// server, and unrecognized values are already rejected by the choice
	// validator on rule_type.
	if ruleType.IsNull() || ruleType.IsUnknown() {
		return
	}
	expectedAttr, recognized := ruleTypeToConfigAttr[ruleType.ValueString()]
	if !recognized {
		return
	}
	if (expectedAttr == "http" && !httpSet) ||
		(expectedAttr == "multistep" && !multistepSet) {
		resp.Diagnostics.AddAttributeError(
			path.Root("rule_config"),
			"Mismatched rule configuration",
			fmt.Sprintf(
				"rule_type is %q but rule_config.%s is not set.",
				ruleType.ValueString(), expectedAttr,
			),
		)
	}
}
