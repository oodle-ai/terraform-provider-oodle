package validatorutils

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"time"
)

type durationValidator struct {
}

func NewDurationValidator() validator.String {
	return &durationValidator{}
}

var _ validator.String = (*durationValidator)(nil)

func (d durationValidator) Description(ctx context.Context) string {
	return "Validates that the string is a valid duration"
}

func (d durationValidator) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d durationValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	_, err := time.ParseDuration(request.ConfigValue.String())
	if err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid duration",
			fmt.Sprintf("The value %v is not a valid duration: %v", request.ConfigValue.String(), err))
	}
}
