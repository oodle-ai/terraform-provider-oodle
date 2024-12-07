package validatorutils

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type durationModifier struct {
}

var _ planmodifier.String = (*durationModifier)(nil)

// NewDurationModifier modifies duration based on golang requirements.
// For example, golang requires "3m" to be formatted as "3m0s"
func NewDurationModifier() planmodifier.String {
	return &durationModifier{}
}

func (d durationModifier) Description(ctx context.Context) string {
	return "Format duration based on golang requirements"
}

func (d durationModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d durationModifier) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	if request.ConfigValue.IsNull() {
		return
	}

	dur, err := time.ParseDuration(request.ConfigValue.ValueString())
	if err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			fmt.Sprintf("Invalid duration %v", request.ConfigValue.ValueString()),
			err.Error())
		return
	}

	if dur == 0 {
		return
	}

	response.PlanValue = types.StringValue(dur.String())
}
