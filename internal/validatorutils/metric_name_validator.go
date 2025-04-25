package validatorutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type metricNameValidator struct {
}

var _ validator.String = (*metricNameValidator)(nil)

func NewMetricNameValidator() validator.String {
	return &metricNameValidator{}
}

func (m metricNameValidator) Description(ctx context.Context) string {
	return "Validates that the metric name starts with oodle_logs_"
}

func (m metricNameValidator) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m metricNameValidator) ValidateString(
	ctx context.Context,
	request validator.StringRequest,
	response *validator.StringResponse,
) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	if !strings.HasPrefix(value, "oodle_logs_") {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid metric name",
			fmt.Sprintf("Metric name must start with oodle_logs_, got: %s", value),
		)
	}
}
