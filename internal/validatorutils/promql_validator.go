package validatorutils

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"go/parser"
)

type promqlValidator struct {
}

var _ validator.String = (*promqlValidator)(nil)

func NewPromQLValidator() validator.String {
	return &promqlValidator{}
}

func (p promqlValidator) Description(ctx context.Context) string {
	return "Validates that the string is a valid promql query"
}

func (p promqlValidator) MarkdownDescription(ctx context.Context) string {
	return p.Description(ctx)
}

func (p promqlValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	_, err := parser.ParseExpr(request.ConfigValue.String())
	if err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid promql query",
			fmt.Sprintf(
				"The value %v is not a valid promql query: %v",
				request.ConfigValue.String(),
				err,
			),
		)
	}
}
