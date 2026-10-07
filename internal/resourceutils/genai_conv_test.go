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

func stringList(t *testing.T, values ...string) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), types.StringType, values)
	if diags.HasError() {
		t.Fatalf("list: %v", diags)
	}
	return list
}

func listValues(t *testing.T, list types.List) []string {
	t.Helper()
	values, err := StringListToSlice(context.Background(), list)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	return values
}

// The server's order is not a change. Taken as it comes, every plan
// would show the list moving and apply it again.
func TestUnorderedSliceToStringListKeepsPriorOrder(t *testing.T) {
	var diags diag.Diagnostics
	prior := stringList(t, "shopassist", "demo")

	got := UnorderedSliceToStringList(
		context.Background(), []string{"demo", "shopassist"}, prior, &diags,
	)

	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if !got.Equal(prior) {
		t.Fatalf("got %v, want the prior order %v", got, prior)
	}
}

func TestUnorderedSliceToStringListTakesARealChange(t *testing.T) {
	var diags diag.Diagnostics
	prior := stringList(t, "a", "b")

	for _, server := range [][]string{
		{"a", "c"},
		{"a"},
		{"a", "b", "b"},
	} {
		got := UnorderedSliceToStringList(
			context.Background(), server, prior, &diags,
		)
		values := listValues(t, got)
		if len(values) != len(server) {
			t.Fatalf("server %v: got %v", server, values)
		}
		for i := range server {
			if values[i] != server[i] {
				t.Fatalf("server %v: got %v", server, values)
			}
		}
	}
}

func TestUnorderedSliceToStringListNullPrior(t *testing.T) {
	var diags diag.Diagnostics

	got := UnorderedSliceToStringList(
		context.Background(), nil, types.ListNull(types.StringType), &diags,
	)
	if !got.IsNull() {
		t.Fatalf("an unset list with no values must stay null, got %v", got)
	}

	got = UnorderedSliceToStringList(
		context.Background(), []string{"x"}, types.ListNull(types.StringType), &diags,
	)
	if values := listValues(t, got); len(values) != 1 || values[0] != "x" {
		t.Fatalf("got %v", values)
	}
}

func TestSameStrings(t *testing.T) {
	tests := []struct {
		a, b []string
		want bool
	}{
		{[]string{"a", "b"}, []string{"b", "a"}, true},
		{nil, []string{}, true},
		{[]string{"a"}, []string{"b"}, false},
		{[]string{"a", "a"}, []string{"a", "b"}, false},
		{[]string{"a", "b"}, []string{"a"}, false},
	}
	for _, tt := range tests {
		if got := SameStrings(tt.a, tt.b); got != tt.want {
			t.Errorf("SameStrings(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestSameStringLists(t *testing.T) {
	ctx := context.Background()
	if !SameStringLists(ctx, stringList(t, "x", "y"), stringList(t, "y", "x")) {
		t.Error("reordered lists must be the same")
	}
	if SameStringLists(ctx, stringList(t, "x"), stringList(t, "y")) {
		t.Error("different lists must differ")
	}
	if !SameStringLists(ctx, types.ListNull(types.StringType), stringList(t)) {
		t.Error("null and empty both hold no strings")
	}
	if SameStringLists(ctx, types.ListUnknown(types.StringType), stringList(t)) {
		t.Error("an unknown list is not known to be empty")
	}
}
