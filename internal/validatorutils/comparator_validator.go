package validatorutils

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type comparatorValidator struct {
}

var _ validator.String = (*comparatorValidator)(nil)

func NewComparatorValidator() validator.String {
	return &comparatorValidator{}
}

func (c comparatorValidator) Description(ctx context.Context) string {
	return "Validates that the string is a valid comparator like '==', '!=', '>', '<', '>=', '<='"
}

func (c comparatorValidator) MarkdownDescription(ctx context.Context) string {
	return c.Description(ctx)
}

func (c comparatorValidator) ValidateString(
	ctx context.Context,
	request validator.StringRequest,
	response *validator.StringResponse,
) {
	if request.ConfigValue.IsNull() {
		return
	}

	switch request.ConfigValue.ValueString() {
	case "==", "!=", ">", "<", ">=", "<=":
		return
	}

	response.Diagnostics.AddAttributeError(
		request.Path,
		"Invalid comparator",
		fmt.Sprintf(
			"The value %v is not a valid comparator like '==', '!=', '>', '<', '>=', '<='",
			request.ConfigValue.ValueString(),
		),
	)
}
