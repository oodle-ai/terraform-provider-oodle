package validatorutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type defaultList struct {
	defaultValue types.List
}

var _ defaults.List = (*defaultList)(nil)

func NewDefaultEmptyStringList() defaults.List {
	var l []attr.Value
	lv, _ := types.ListValue(basetypes.StringType{}, l)
	return NewDefaultList(lv)
}

func NewDefaultList(defaultValue types.List) defaults.List {
	return &defaultList{
		defaultValue: defaultValue,
	}
}

func (d defaultList) Description(ctx context.Context) string {
	return "defaultList is a schema default value for types.List attributes."
}

func (d defaultList) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d defaultList) DefaultList(ctx context.Context, request defaults.ListRequest, response *defaults.ListResponse) {
	response.PlanValue = d.defaultValue
}
