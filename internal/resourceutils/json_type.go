package resourceutils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// JSONType holds a free-form JSON document. It is jsontypes.NormalizedType
// with two allowances, both for backward compatibility with configurations
// written while these attributes were plain strings:
//
//   - An empty string stays valid, and means the attribute is unset. The
//     plain-string conversion treated it that way, thus a configuration that
//     passes a variable which defaults to "" would otherwise stop planning.
//   - Two numbers that differ only in how they are written, such as 1 and 1.0,
//     compare equal. The plain-string comparison decoded both to a float, thus
//     a value the API returns as 1.0 against a configuration that writes 1
//     would otherwise diff on every plan and never converge.
type JSONType struct {
	jsontypes.NormalizedType
}

var (
	_ basetypes.StringTypable     = JSONType{}
	_ basetypes.StringValuable    = JSON{}
	_ xattr.ValidateableAttribute = JSON{}
)

func (t JSONType) String() string {
	return "resourceutils.JSONType"
}

func (t JSONType) Equal(o attr.Type) bool {
	other, ok := o.(JSONType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t JSONType) ValueType(_ context.Context) attr.Value {
	return JSON{}
}

func (t JSONType) ValueFromString(
	_ context.Context,
	in basetypes.StringValue,
) (basetypes.StringValuable, diag.Diagnostics) {
	return JSON{Normalized: jsontypes.Normalized{StringValue: in}}, nil
}

func (t JSONType) ValueFromTerraform(
	ctx context.Context,
	in tftypes.Value,
) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	value, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to read the JSON value: %v", diags)
	}

	return value, nil
}

// JSON is the value of a JSONType attribute.
type JSON struct {
	jsontypes.Normalized
}

func (v JSON) Type(_ context.Context) attr.Type {
	return JSONType{}
}

func (v JSON) Equal(o attr.Value) bool {
	other, ok := o.(JSON)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// ValidateAttribute accepts an empty string, which means the attribute is
// unset, and otherwise checks that the string is JSON.
func (v JSON) ValidateAttribute(
	ctx context.Context,
	req xattr.ValidateAttributeRequest,
	resp *xattr.ValidateAttributeResponse,
) {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return
	}

	v.Normalized.ValidateAttribute(ctx, req, resp)
}

// StringSemanticEquals compares two documents by value, and reads a number by
// what it holds rather than by how it is written.
func (v JSON) StringSemanticEquals(
	_ context.Context,
	newValuable basetypes.StringValuable,
) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(JSON)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic "+
				"equality checks. Please report this to the provider "+
				"developers.",
		)

		return false, diags
	}

	// An empty string means unset, thus it equals only another empty string.
	if v.ValueString() == "" || newValue.ValueString() == "" {
		return v.ValueString() == newValue.ValueString(), diags
	}

	return jsonEquivalent(
		[]byte(v.ValueString()),
		[]byte(newValue.ValueString()),
	), diags
}

// NewJSONNull returns a null JSON value.
func NewJSONNull() JSON {
	return JSON{Normalized: jsontypes.NewNormalizedNull()}
}

// NewJSONUnknown returns an unknown JSON value.
func NewJSONUnknown() JSON {
	return JSON{Normalized: jsontypes.NewNormalizedUnknown()}
}

// NewJSONValue returns a JSON value holding the given document.
func NewJSONValue(value string) JSON {
	return JSON{Normalized: jsontypes.NewNormalizedValue(value)}
}
