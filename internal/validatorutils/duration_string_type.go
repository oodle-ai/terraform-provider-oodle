package validatorutils

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Compile-time interface checks.
var (
	_ basetypes.StringTypable                    = DurationType{}
	_ basetypes.StringValuableWithSemanticEquals = DurationValue{}
)

// DurationType is a custom Terraform Framework type that represents a duration string.
// It implements basetypes.StringTypable.
type DurationType struct {
	basetypes.StringType
}

// NewDurationType returns a new DurationType.
func NewDurationType() DurationType {
	return DurationType{}
}

// Equal returns true if the given type is a DurationType.
func (t DurationType) Equal(o attr.Type) bool {
	_, ok := o.(DurationType)
	return ok
}

// String returns a human-readable string of the type name.
func (t DurationType) String() string {
	return "DurationType"
}

// ValueFromString wraps a StringValue in a DurationValue.
func (t DurationType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return DurationValue{StringValue: in}, nil
}

// ValueFromTerraform converts a tftypes.Value into a DurationValue.
func (t DurationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

// ValueType returns the value type of this type.
func (t DurationType) ValueType(ctx context.Context) attr.Value {
	return DurationValue{}
}

// DurationValue is a custom Terraform Framework value that represents a duration string.
// It implements basetypes.StringValuableWithSemanticEquals to allow semantic comparison
// of duration strings (e.g., "0.75m" == "45s").
type DurationValue struct {
	basetypes.StringValue
}

// NewDurationValue returns a new DurationValue with the given string.
func NewDurationValue(s string) DurationValue {
	return DurationValue{StringValue: basetypes.NewStringValue(s)}
}

// NewDurationNull returns a new null DurationValue.
func NewDurationNull() DurationValue {
	return DurationValue{StringValue: basetypes.NewStringNull()}
}

// NewDurationUnknown returns a new unknown DurationValue.
// Provided for completeness alongside NewDurationValue and NewDurationNull as part of the
// DurationValue type API. Currently unused but available for future use.
func NewDurationUnknown() DurationValue {
	return DurationValue{StringValue: basetypes.NewStringUnknown()}
}

// Equal returns true if the given value is a DurationValue with the same underlying string.
func (v DurationValue) Equal(o attr.Value) bool {
	other, ok := o.(DurationValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

// Type returns the type of this value.
func (v DurationValue) Type(ctx context.Context) attr.Type {
	return DurationType{}
}

// ToStringValue returns the underlying StringValue.
func (v DurationValue) ToStringValue(ctx context.Context) (basetypes.StringValue, diag.Diagnostics) {
	return v.StringValue, nil
}

// StringSemanticEquals compares two duration strings semantically by parsing them
// as time.Duration values. For example, "0.75m" and "45s" are semantically equal.
//
// When either value fails to parse as a duration, this method returns (false, nil)
// intentionally — unparseable strings cannot be semantically equal, and parse errors
// are expected to be caught by the DurationValidator during validation. Returning
// false without diagnostics ensures Terraform treats the values as different without
// producing confusing error messages during the semantic equality check phase.
func (v DurationValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	newStringValue, diags := newValuable.ToStringValue(ctx)
	if diags.HasError() {
		return false, diags
	}

	priorDuration, err := time.ParseDuration(v.ValueString())
	if err != nil {
		// Prior value is unparseable — cannot be semantically equal.
		return false, nil
	}

	newDuration, err := time.ParseDuration(newStringValue.ValueString())
	if err != nil {
		// New value is unparseable — cannot be semantically equal.
		return false, nil
	}

	return priorDuration == newDuration, nil
}
