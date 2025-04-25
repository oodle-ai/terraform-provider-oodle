package validatorutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type filterValidator struct {
}

var _ validator.Object = (*filterValidator)(nil)

func NewFilterValidator() validator.Object {
	return &filterValidator{}
}

func (f filterValidator) Description(ctx context.Context) string {
	return "Validates that only one of match/all/any/not is specified in a filter"
}

func (f filterValidator) MarkdownDescription(ctx context.Context) string {
	return f.Description(ctx)
}

func (f filterValidator) ValidateObject(
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
			"Missing filter attributes",
			"The filter must contain at least one of match, all, any, or not",
		)
		return
	}

	// Count how many are set (not null)
	count := 0

	// Check each field
	if matchAttr, ok := attrs["match"].(attr.Value); ok && !matchAttr.IsNull() {
		count++
	}
	if allAttr, ok := attrs["all"].(attr.Value); ok && !allAttr.IsNull() {
		count++
	}
	if anyAttr, ok := attrs["any"].(attr.Value); ok && !anyAttr.IsNull() {
		count++
	}
	if notAttr, ok := attrs["not"].(attr.Value); ok && !notAttr.IsNull() {
		count++
	}

	if count == 0 {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"No filter type specified",
			"Exactly one of match, all, any, or not must be specified in a filter",
		)
		return
	}

	if count > 1 {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Multiple filter types specified",
			"Only one of match, all, any, or not can be specified in a filter",
		)
	}
}
