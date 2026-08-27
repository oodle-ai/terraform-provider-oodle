package validatorutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// syntheticMonitorConfigValidator validates, at plan time, that exactly one
// rule_config block is set and that it agrees with rule_type. Without this the
// mismatch is only caught by the server on apply, producing a worse UX.
type syntheticMonitorConfigValidator struct{}

var _ resource.ConfigValidator = (*syntheticMonitorConfigValidator)(nil)

// syntheticMonitorRuleConfigAttrs are the rule_config attribute names, one per
// rule_type. Each rule_type maps to the identically named attribute, so this
// doubles as the set of recognized rule types. Listed in schema order to keep
// error messages stable.
var syntheticMonitorRuleConfigAttrs = []string{
	"http",
	"ping",
	"dns",
	"tcp",
	"traceroute",
	"ssl",
	"multistep",
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

// isRecognizedRuleType reports whether ruleType names a known rule_config
// block. Unrecognized values are the choice validator's to report, so this
// validator stays silent on them rather than emitting a confusing second error.
func isRecognizedRuleType(ruleType string) bool {
	for _, attr := range syntheticMonitorRuleConfigAttrs {
		if attr == ruleType {
			return true
		}
	}
	return false
}

// quotedRuleConfigAttrs renders the valid block names for error messages, e.g.
// `rule_config.http, rule_config.ping, ...`.
func quotedRuleConfigAttrs() string {
	names := make([]string, len(syntheticMonitorRuleConfigAttrs))
	for i, attr := range syntheticMonitorRuleConfigAttrs {
		names[i] = "rule_config." + attr
	}
	return strings.Join(names, ", ")
}

func (v syntheticMonitorConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var ruleType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx,
		path.Root("rule_type"), &ruleType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Collect the names of every rule_config block that is set.
	var setAttrs []string
	for _, attr := range syntheticMonitorRuleConfigAttrs {
		var config types.Object
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx,
			path.Root("rule_config").AtName(attr), &config)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !config.IsNull() && !config.IsUnknown() {
			setAttrs = append(setAttrs, attr)
		}
	}

	// Exactly one rule_config block must be set.
	switch len(setAttrs) {
	case 1:
	case 0:
		resp.Diagnostics.AddAttributeError(
			path.Root("rule_config"),
			"Missing rule configuration",
			fmt.Sprintf(
				"Exactly one of %s must be set.",
				quotedRuleConfigAttrs(),
			),
		)
		return
	default:
		resp.Diagnostics.AddAttributeError(
			path.Root("rule_config"),
			"Conflicting rule configuration",
			fmt.Sprintf(
				"Only one rule_config block may be set, but %d are set: %s.",
				len(setAttrs), strings.Join(setAttrs, ", "),
			),
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
	if !isRecognizedRuleType(ruleType.ValueString()) {
		return
	}
	if setAttrs[0] != ruleType.ValueString() {
		resp.Diagnostics.AddAttributeError(
			path.Root("rule_config"),
			"Mismatched rule configuration",
			fmt.Sprintf(
				"rule_type is %q but rule_config.%s is not set (rule_config.%s is set instead).",
				ruleType.ValueString(), ruleType.ValueString(), setAttrs[0],
			),
		)
	}
}
