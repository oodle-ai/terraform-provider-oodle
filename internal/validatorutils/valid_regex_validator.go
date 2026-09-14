package validatorutils

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// validRegexValidator checks that the attribute's own value compiles as a
// regular expression. This is the inverse of NewRegexValidator, which matches
// a value against a fixed pattern.
type validRegexValidator struct {
}

var _ validator.String = (*validRegexValidator)(nil)

// NewValidRegexValidator validates that the configured value is a syntactically
// valid regular expression. Terraform and the Oodle backend both use Go's RE2
// engine, so a pattern accepted here is accepted server-side too.
func NewValidRegexValidator() validator.String {
	return &validRegexValidator{}
}

func (v validRegexValidator) Description(_ context.Context) string {
	return "Validates that the value is a valid regular expression"
}

func (v validRegexValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validRegexValidator) ValidateString(
	_ context.Context,
	request validator.StringRequest,
	response *validator.StringResponse,
) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	if _, err := regexp.Compile(value); err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid regular expression",
			fmt.Sprintf("Value must be a valid regular expression, got %q: %v", value, err),
		)
	}
}
