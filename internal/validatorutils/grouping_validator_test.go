package validatorutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"
)

func TestGroupingValidator(t *testing.T) {
	v := NewGroupingValidator()
	ctx := context.Background()

	// Test valid cases - exactly one field set
	testCases := []struct {
		name        string
		attrs       map[string]attr.Value
		shouldError bool
	}{
		{
			name: "disabled only",
			attrs: map[string]attr.Value{
				"disabled":   types.BoolValue(true),
				"by_monitor": types.BoolNull(),
				"by_labels":  types.ListNull(types.StringType),
			},
			shouldError: false,
		},
		{
			name: "by_monitor only",
			attrs: map[string]attr.Value{
				"disabled":   types.BoolNull(),
				"by_monitor": types.BoolValue(true),
				"by_labels":  types.ListNull(types.StringType),
			},
			shouldError: false,
		},
		{
			name: "by_labels only",
			attrs: map[string]attr.Value{
				"disabled":   types.BoolNull(),
				"by_monitor": types.BoolNull(),
				"by_labels": types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("service"),
					types.StringValue("region"),
				}),
			},
			shouldError: false,
		},
		{
			name: "multiple fields set",
			attrs: map[string]attr.Value{
				"disabled":   types.BoolValue(true),
				"by_monitor": types.BoolValue(true),
				"by_labels":  types.ListNull(types.StringType),
			},
			shouldError: true,
		},
		{
			name: "all fields set",
			attrs: map[string]attr.Value{
				"disabled":   types.BoolValue(true),
				"by_monitor": types.BoolValue(true),
				"by_labels": types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("service"),
				}),
			},
			shouldError: true,
		},
		{
			name: "no fields set",
			attrs: map[string]attr.Value{
				"disabled":   types.BoolNull(),
				"by_monitor": types.BoolNull(),
				"by_labels":  types.ListNull(types.StringType),
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objType := types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"disabled":   types.BoolType,
					"by_monitor": types.BoolType,
					"by_labels":  types.ListType{ElemType: types.StringType},
				},
			}

			objValue, diags := types.ObjectValue(objType.AttrTypes, tc.attrs)
			assert.False(t, diags.HasError(), "Failed to create test object: %v", diags)

			req := validator.ObjectRequest{
				ConfigValue: objValue,
			}
			resp := &validator.ObjectResponse{}

			v.ValidateObject(ctx, req, resp)

			if tc.shouldError {
				assert.True(t, resp.Diagnostics.HasError(), "Expected validation error for case: %s", tc.name)
			} else {
				assert.False(t, resp.Diagnostics.HasError(), "Unexpected validation error for case: %s, errors: %v", tc.name, resp.Diagnostics)
			}
		})
	}

	// Test null object (should pass)
	req := validator.ObjectRequest{
		ConfigValue: types.ObjectNull(map[string]attr.Type{
			"disabled":   types.BoolType,
			"by_monitor": types.BoolType,
			"by_labels":  types.ListType{ElemType: types.StringType},
		}),
	}
	resp := &validator.ObjectResponse{}
	v.ValidateObject(ctx, req, resp)
	assert.False(t, resp.Diagnostics.HasError(), "Null object should not produce validation errors")

	// Test unknown object (should pass)
	req = validator.ObjectRequest{
		ConfigValue: types.ObjectUnknown(map[string]attr.Type{
			"disabled":   types.BoolType,
			"by_monitor": types.BoolType,
			"by_labels":  types.ListType{ElemType: types.StringType},
		}),
	}
	resp = &validator.ObjectResponse{}
	v.ValidateObject(ctx, req, resp)
	assert.False(t, resp.Diagnostics.HasError(), "Unknown object should not produce validation errors")
}
