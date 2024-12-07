package validatorutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type choiceValidator struct {
	validChoices map[string]struct{}
}

var _ validator.String = (*choiceValidator)(nil)

func NewChoiceValidator(choices map[string]struct{}) validator.String {
	return &choiceValidator{
		validChoices: choices,
	}
}

func (c choiceValidator) choiceRepr() string {
	var validComparatorsList []string
	for k := range c.validChoices {
		validComparatorsList = append(validComparatorsList, k)
	}

	return strings.Join(validComparatorsList, ", ")
}

func (c choiceValidator) Description(ctx context.Context) string {
	return "Validates that the string is a one of " + c.choiceRepr()
}

func (c choiceValidator) MarkdownDescription(ctx context.Context) string {
	return c.Description(ctx)
}

func (c choiceValidator) ValidateString(
	ctx context.Context,
	request validator.StringRequest,
	response *validator.StringResponse,
) {
	if request.ConfigValue.IsNull() {
		return
	}

	_, ok := c.validChoices[request.ConfigValue.ValueString()]
	if ok {
		return
	}

	var validComparatorsList []string
	for k := range c.validChoices {
		validComparatorsList = append(validComparatorsList, k)
	}

	response.Diagnostics.AddAttributeError(
		request.Path,
		"Invalid comparator",
		fmt.Sprintf(
			"The value %v is should be one of "+strings.Join(validComparatorsList, ", "),
			request.ConfigValue.ValueString(),
		),
	)
}
