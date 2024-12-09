package validatorutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/rubrikinc/testwell/assert"
)

func TestDefaultList(t *testing.T) {
	ctx := context.TODO()
	var l []attr.Value
	l = append(l, types.StringValue("foo"))
	lv, diags := types.ListValue(basetypes.StringType{}, l)
	assert.False(t, diags.HasError())

	ds := NewDefaultList(lv)

	resp := &defaults.ListResponse{}
	req := defaults.ListRequest{}
	ds.DefaultList(ctx, req, resp)
	assert.DeepEqual(t, lv, resp.PlanValue)

	nilDs := NewDefaultList(types.ListNull(basetypes.StringType{}))
	nilDs.DefaultList(ctx, req, resp)
	assert.True(t, resp.PlanValue.IsNull())
}
