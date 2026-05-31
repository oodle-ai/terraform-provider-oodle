package validatorutils

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type regexValidator struct {
	pattern     *regexp.Regexp
	description string
}

var _ validator.String = (*regexValidator)(nil)

// NewRegexValidator returns a string validator that fails when the input
// does not match pattern. The description is shown in plan diagnostics
// when validation fails (e.g. "must be a 12-digit AWS account ID").
func NewRegexValidator(pattern *regexp.Regexp, description string) validator.String {
	return &regexValidator{pattern: pattern, description: description}
}

func (v regexValidator) Description(_ context.Context) string {
	return v.description
}

func (v regexValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v regexValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	if !v.pattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			fmt.Sprintf("Value %q %s", value, v.description),
		)
	}
}
