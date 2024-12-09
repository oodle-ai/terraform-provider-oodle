package validatorutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultString struct {
	defaultValue types.String
}

var _ defaults.String = (*defaultString)(nil)

func NewDefaultString(defaultValue types.String) defaults.String {
	return &defaultString{
		defaultValue: defaultValue,
	}
}

func (d defaultString) Description(ctx context.Context) string {
	return "defaultString is a schema default value for types.String attributes."
}

func (d defaultString) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d defaultString) DefaultString(ctx context.Context, request defaults.StringRequest, response *defaults.StringResponse) {
	response.PlanValue = d.defaultValue
}
