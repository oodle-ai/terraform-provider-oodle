package validatorutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type groupingValidator struct {
}

var _ validator.Object = (*groupingValidator)(nil)

func NewGroupingValidator() validator.Object {
	return &groupingValidator{}
}

func (g groupingValidator) Description(ctx context.Context) string {
	return "Validates that exactly one of by_monitor, by_labels, or disabled is specified in grouping"
}

func (g groupingValidator) MarkdownDescription(ctx context.Context) string {
	return g.Description(ctx)
}

func (g groupingValidator) ValidateObject(
	ctx context.Context,
	request validator.ObjectRequest,
	response *validator.ObjectResponse,
) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	attrs := request.ConfigValue.Attributes()
	if attrs == nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Missing grouping attributes",
			"The grouping must contain exactly one of by_monitor, by_labels, or disabled",
		)
		return
	}

	// Count how many are set (not null)
	count := 0

	// Check each field
	if byMonitorAttr, ok := attrs["by_monitor"]; ok && !byMonitorAttr.IsNull() {
		count++
	}
	if byLabelsAttr, ok := attrs["by_labels"]; ok && !byLabelsAttr.IsNull() {
		count++
	}
	if disabledAttr, ok := attrs["disabled"]; ok && !disabledAttr.IsNull() {
		count++
	}

	if count == 0 {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"No grouping type specified",
			"Exactly one of by_monitor, by_labels, or disabled must be specified in grouping",
		)
		return
	}

	if count > 1 {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Multiple grouping types specified",
			"Only one of by_monitor, by_labels, or disabled can be specified in grouping",
		)
	}
}
