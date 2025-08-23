package validatorutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNotifiersValidator(t *testing.T) {
	ctx := context.Background()
	v := NewNotifiersValidator()

	tests := []struct {
		name        string
		input       types.Object
		expectError bool
	}{
		{
			name:        "null value should pass",
			input:       types.ObjectNull(nil),
			expectError: false,
		},
		{
			name: "any with empty lists should pass",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"any":      types.ListType{ElemType: types.StringType},
					"warn":     types.ListType{ElemType: types.StringType},
					"critical": types.ListType{ElemType: types.StringType},
					"no_data":  types.ListType{ElemType: types.StringType},
				},
				map[string]attr.Value{
					"any":      types.ListValueMust(types.StringType, []attr.Value{types.StringValue("notifier1")}),
					"warn":     types.ListNull(types.StringType),
					"critical": types.ListNull(types.StringType),
					"no_data":  types.ListNull(types.StringType),
				},
			),
			expectError: false,
		},
		{
			name: "any with warn should fail",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"any":      types.ListType{ElemType: types.StringType},
					"warn":     types.ListType{ElemType: types.StringType},
					"critical": types.ListType{ElemType: types.StringType},
					"no_data":  types.ListType{ElemType: types.StringType},
				},
				map[string]attr.Value{
					"any":      types.ListValueMust(types.StringType, []attr.Value{types.StringValue("notifier1")}),
					"warn":     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("notifier2")}),
					"critical": types.ListNull(types.StringType),
					"no_data":  types.ListNull(types.StringType),
				},
			),
			expectError: true,
		},
		{
			name: "warn and critical without any should pass",
			input: types.ObjectValueMust(
				map[string]attr.Type{
					"any":      types.ListType{ElemType: types.StringType},
					"warn":     types.ListType{ElemType: types.StringType},
					"critical": types.ListType{ElemType: types.StringType},
					"no_data":  types.ListType{ElemType: types.StringType},
				},
				map[string]attr.Value{
					"any":      types.ListNull(types.StringType),
					"warn":     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("notifier1")}),
					"critical": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("notifier2")}),
					"no_data":  types.ListNull(types.StringType),
				},
			),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.ObjectRequest{
				ConfigValue: tt.input,
			}
			resp := &validator.ObjectResponse{}

			v.ValidateObject(ctx, req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.expectError {
				t.Errorf("expected error: %v, got error: %v, diagnostics: %v",
					tt.expectError, hasError, resp.Diagnostics)
			}
		})
	}
}
