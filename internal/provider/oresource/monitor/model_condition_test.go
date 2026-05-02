package monitor

import (
	"testing"
	"time"

	"terraform-provider-oodle/internal/validatorutils"
)

func TestDurationValueFromModel(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		wantNull bool
		wantVal  string
	}{
		{
			name:     "positive duration returns value",
			duration: 5 * time.Minute,
			wantNull: false,
			wantVal:  "5m",
		},
		{
			name:     "zero duration returns null",
			duration: 0,
			wantNull: true,
		},
		{
			name:     "sub-second duration returns value",
			duration: 500 * time.Millisecond,
			wantNull: false,
			wantVal:  "500ms",
		},
		{
			name:     "complex duration returns value",
			duration: 2*time.Hour + 30*time.Minute,
			wantNull: false,
			wantVal:  "2h30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := durationValueFromModel(tt.duration)
			if tt.wantNull {
				if !result.IsNull() {
					t.Errorf("expected null, got %q", result.ValueString())
				}
			} else {
				if result.IsNull() {
					t.Error("expected non-null value, got null")
				}
				if result.ValueString() != tt.wantVal {
					t.Errorf("expected %q, got %q", tt.wantVal, result.ValueString())
				}
			}
		})
	}
}

func TestDurationValueFromDurationPtr(t *testing.T) {
	tests := []struct {
		name     string
		duration *time.Duration
		wantNull bool
		wantVal  string
	}{
		{
			name:     "nil pointer returns null",
			duration: nil,
			wantNull: true,
		},
		{
			name:     "non-nil positive duration returns value",
			duration: durationPtr(5 * time.Minute),
			wantNull: false,
			wantVal:  "5m",
		},
		{
			name:     "non-nil zero duration returns value",
			duration: durationPtr(0),
			wantNull: false,
			wantVal:  "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := durationValueFromDurationPtr(tt.duration)
			if tt.wantNull {
				if !result.IsNull() {
					t.Errorf("expected null, got %q", result.ValueString())
				}
			} else {
				if result.IsNull() {
					t.Error("expected non-null value, got null")
				}
				if result.ValueString() != tt.wantVal {
					t.Errorf("expected %q, got %q", tt.wantVal, result.ValueString())
				}
			}
		})
	}
}

func TestParseDurationValue(t *testing.T) {
	tests := []struct {
		name      string
		value     validatorutils.DurationValue
		fieldName string
		wantDur   time.Duration
		wantErr   bool
	}{
		{
			name:      "valid duration",
			value:     validatorutils.NewDurationValue("5m"),
			fieldName: "test_field",
			wantDur:   5 * time.Minute,
			wantErr:   false,
		},
		{
			name:      "null value returns zero",
			value:     validatorutils.NewDurationNull(),
			fieldName: "test_field",
			wantDur:   0,
			wantErr:   false,
		},
		{
			name:      "unknown value returns zero",
			value:     validatorutils.NewDurationUnknown(),
			fieldName: "test_field",
			wantDur:   0,
			wantErr:   false,
		},
		{
			name:      "zero duration string",
			value:     validatorutils.NewDurationValue("0s"),
			fieldName: "test_field",
			wantDur:   0,
			wantErr:   false,
		},
		{
			name:      "invalid duration returns error",
			value:     validatorutils.NewDurationValue("invalid"),
			fieldName: "test_field",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDurationValue(tt.value, tt.fieldName)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.wantDur {
					t.Errorf("expected %v, got %v", tt.wantDur, result)
				}
			}
		})
	}
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
