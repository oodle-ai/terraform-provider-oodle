package validatorutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"
)

func TestAliasValidatorRefusesBothNames(t *testing.T) {
	listType := types.ListType{ElemType: types.StringType}
	one := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("x")})
	null := types.ListNull(types.StringType)
	obj := func(warn, warning attr.Value) types.Object {
		return types.ObjectValueMust(
			map[string]attr.Type{"warn": listType, "warning": listType},
			map[string]attr.Value{"warn": warn, "warning": warning},
		)
	}
	check := func(o types.Object) bool {
		resp := &validator.ObjectResponse{}
		NewAliasValidator("warn", "warning").ValidateObject(
			context.Background(), validator.ObjectRequest{ConfigValue: o}, resp,
		)
		return resp.Diagnostics.HasError()
	}
	assert.False(t, check(obj(one, null)))
	assert.False(t, check(obj(null, one)))
	assert.False(t, check(obj(null, null)))
	assert.True(t, check(obj(one, one)))
}
