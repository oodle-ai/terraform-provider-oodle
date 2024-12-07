package validatorutils

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type uuidValidator struct {
}

var _ validator.String = (*uuidValidator)(nil)

func NewUUIDValidator() validator.String {
	return &uuidValidator{}
}

func (u uuidValidator) Description(ctx context.Context) string {
	return "Validates that the string is a valid UUID"
}

func (u uuidValidator) MarkdownDescription(ctx context.Context) string {
	return u.Description(ctx)
}

func (u uuidValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if err := uuid.Validate(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			fmt.Sprintf("Invalid UUID: %v", request.ConfigValue.ValueString()),
			err.Error(),
		)
	}
}
