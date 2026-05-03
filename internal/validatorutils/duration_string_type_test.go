package validatorutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStringSemanticEquals(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		valueA   string
		valueB   string
		expected bool
	}{
		{
			name:     "0.75m equals 45s",
			valueA:   "0.75m",
			valueB:   "45s",
			expected: true,
		},
		{
			name:     "1.5m equals 1m30s",
			valueA:   "1.5m",
			valueB:   "1m30s",
			expected: true,
		},
		{
			name:     "0.5m equals 30s",
			valueA:   "0.5m",
			valueB:   "30s",
			expected: true,
		},
		{
			name:     "2.5h equals 2h30m",
			valueA:   "2.5h",
			valueB:   "2h30m",
			expected: true,
		},
		{
			name:     "5m equals 5m identical",
			valueA:   "5m",
			valueB:   "5m",
			expected: true,
		},
		{
			name:     "5m does not equal 10m",
			valueA:   "5m",
			valueB:   "10m",
			expected: false,
		},
		{
			name:     "1s does not equal 1m",
			valueA:   "1s",
			valueB:   "1m",
			expected: false,
		},
		{
			name:     "invalid prior value returns false",
			valueA:   "abc",
			valueB:   "5m",
			expected: false,
		},
		{
			name:     "invalid new value returns false",
			valueA:   "5m",
			valueB:   "abc",
			expected: false,
		},
		{
			name:     "both values invalid returns false",
			valueA:   "abc",
			valueB:   "xyz",
			expected: false,
		},
		{
			name:     "empty prior value returns false",
			valueA:   "",
			valueB:   "5m",
			expected: false,
		},
		{
			name:     "empty new value returns false",
			valueA:   "5m",
			valueB:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valA := NewDurationValue(tt.valueA)
			valB := NewDurationValue(tt.valueB)

			result, diags := valA.StringSemanticEquals(ctx, valB)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if result != tt.expected {
				t.Errorf("StringSemanticEquals(%q, %q) = %v, want %v", tt.valueA, tt.valueB, result, tt.expected)
			}
		})
	}
}

func TestValueFromTerraform(t *testing.T) {
	ctx := context.Background()
	dt := NewDurationType()

	t.Run("known value", func(t *testing.T) {
		tfVal := tftypes.NewValue(tftypes.String, "5m")
		val, err := dt.ValueFromTerraform(ctx, tfVal)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		dv, ok := val.(DurationValue)
		if !ok {
			t.Fatalf("expected DurationValue, got %T", val)
		}
		if dv.ValueString() != "5m" {
			t.Errorf("expected %q, got %q", "5m", dv.ValueString())
		}
		if dv.IsNull() {
			t.Error("expected non-null value")
		}
		if dv.IsUnknown() {
			t.Error("expected known value")
		}
	})

	t.Run("null value", func(t *testing.T) {
		tfVal := tftypes.NewValue(tftypes.String, nil)
		val, err := dt.ValueFromTerraform(ctx, tfVal)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		dv, ok := val.(DurationValue)
		if !ok {
			t.Fatalf("expected DurationValue, got %T", val)
		}
		if !dv.IsNull() {
			t.Error("expected null value")
		}
	})

	t.Run("unknown value", func(t *testing.T) {
		tfVal := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
		val, err := dt.ValueFromTerraform(ctx, tfVal)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		dv, ok := val.(DurationValue)
		if !ok {
			t.Fatalf("expected DurationValue, got %T", val)
		}
		if !dv.IsUnknown() {
			t.Error("expected unknown value")
		}
	})
}

func TestDurationValueEqual(t *testing.T) {
	t.Run("same value returns true", func(t *testing.T) {
		a := NewDurationValue("5m")
		b := NewDurationValue("5m")
		if !a.Equal(b) {
			t.Error("expected Equal to return true for same values")
		}
	})

	t.Run("different value returns false", func(t *testing.T) {
		a := NewDurationValue("5m")
		b := NewDurationValue("10m")
		if a.Equal(b) {
			t.Error("expected Equal to return false for different values")
		}
	})

	t.Run("different type returns false", func(t *testing.T) {
		a := NewDurationValue("5m")
		b := types.StringValue("5m")
		if a.Equal(b) {
			t.Error("expected Equal to return false for different types")
		}
	})
}

func TestDurationValueType(t *testing.T) {
	ctx := context.Background()
	v := NewDurationValue("5m")
	typ := v.Type(ctx)
	if _, ok := typ.(DurationType); !ok {
		t.Errorf("expected DurationType, got %T", typ)
	}
}
