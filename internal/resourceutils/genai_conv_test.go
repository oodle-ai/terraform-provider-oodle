package resourceutils

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestJSONToRaw(t *testing.T) {
	tests := []struct {
		name    string
		input   JSON
		want    string
		wantErr bool
	}{
		{
			name:  "null becomes nil",
			input: NewJSONNull(),
			want:  "",
		},
		{
			name:  "unknown becomes nil",
			input: NewJSONUnknown(),
			want:  "",
		},
		{
			name:  "empty becomes nil",
			input: NewJSONValue(""),
			want:  "",
		},
		{
			name:  "object passes through",
			input: NewJSONValue(`{"a":1}`),
			want:  `{"a":1}`,
		},
		{
			name:    "invalid JSON is rejected",
			input:   NewJSONValue(`{not json`),
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := JSONToRaw(test.input, "attr")
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", string(got))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != test.want {
				t.Errorf("got %q, want %q", string(got), test.want)
			}
		})
	}
}

func TestRawToJSONIsNullWhenEmpty(t *testing.T) {
	if got := RawToJSON(nil); !got.IsNull() {
		t.Fatalf("expected null, got %q", got.ValueString())
	}

	if got := RawToJSON(json.RawMessage("null")); !got.IsNull() {
		t.Fatalf("expected null, got %q", got.ValueString())
	}

	got := RawToJSON(json.RawMessage(`{"a":1}`))
	if got.ValueString() != `{"a":1}` {
		t.Fatalf("unexpected value %q", got.ValueString())
	}
}

func TestRawToJSONStringKeepsPriorWhenEquivalent(t *testing.T) {
	// The API re-serializes JSON, so key order and whitespace differ
	// from what the user wrote. The configured text has to survive or
	// every plan shows a diff for a value that never changed.
	prior := types.StringValue("{\n  \"b\": 2,\n  \"a\": 1\n}")
	got := RawToJSONString(json.RawMessage(`{"a":1,"b":2}`), prior)

	if got != prior {
		t.Errorf("got %q, want the prior value %q", got, prior)
	}
}

func TestRawToJSONStringTakesServerValueWhenChanged(t *testing.T) {
	prior := types.StringValue(`{"a":1}`)
	got := RawToJSONString(json.RawMessage(`{"a":2}`), prior)

	if got.ValueString() != `{"a":2}` {
		t.Errorf("got %q, want the server value", got)
	}
}

func TestRawToJSONStringEmptyStaysNull(t *testing.T) {
	got := RawToJSONString(nil, types.StringNull())
	if !got.IsNull() {
		t.Errorf("got %q, want null", got)
	}

	got = RawToJSONString(json.RawMessage("null"), types.StringNull())
	if !got.IsNull() {
		t.Errorf("got %q, want null for a JSON null", got)
	}
}

func TestStringListRoundTrip(t *testing.T) {
	ctx := context.Background()
	diagnostics := &diag.Diagnostics{}

	list := SliceToStringList(
		ctx, []string{"a", "b"}, types.ListNull(types.StringType), diagnostics,
	)
	if diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diagnostics.Errors())
	}

	got, err := StringListToSlice(ctx, list)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}

func TestSliceToStringListEmptyStaysNull(t *testing.T) {
	// An optional list the user never set must not start showing a
	// diff against an empty list once the API answers with none.
	diagnostics := &diag.Diagnostics{}
	got := SliceToStringList(
		context.Background(),
		nil,
		types.ListNull(types.StringType),
		diagnostics,
	)

	if !got.IsNull() {
		t.Errorf("got %v, want a null list", got)
	}
}
