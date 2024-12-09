package validatorutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"
)

func TestDefaultString(t *testing.T) {
	ctx := context.TODO()
	ds := NewDefaultString(types.StringValue("foo"))

	resp := &defaults.StringResponse{}
	req := defaults.StringRequest{}
	ds.DefaultString(ctx, req, resp)

	assert.Equal(t, "foo", resp.PlanValue.ValueString())

	nilDs := NewDefaultString(types.StringNull())
	nilDs.DefaultString(ctx, req, resp)
	assert.True(t, resp.PlanValue.IsNull())
}
