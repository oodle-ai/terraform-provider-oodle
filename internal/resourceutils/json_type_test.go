package resourceutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func validateJSON(value JSON) xattr.ValidateAttributeResponse {
	resp := xattr.ValidateAttributeResponse{}
	value.ValidateAttribute(
		context.Background(),
		xattr.ValidateAttributeRequest{Path: path.Root("attr")},
		&resp,
	)

	return resp
}

// An empty string meant "unset" while these attributes were plain strings, thus
// a configuration that holds one has to keep planning.
func TestJSONAcceptsAnEmptyString(t *testing.T) {
	if resp := validateJSON(NewJSONValue("")); resp.Diagnostics.HasError() {
		t.Fatalf("an empty string was rejected: %v", resp.Diagnostics)
	}

	if resp := validateJSON(NewJSONNull()); resp.Diagnostics.HasError() {
		t.Fatalf("a null value was rejected: %v", resp.Diagnostics)
	}
}

func TestJSONRejectsTextThatIsNotJSON(t *testing.T) {
	resp := validateJSON(NewJSONValue(`{not json`))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error for a value that is not JSON")
	}
}

// The plain-string comparison read both numbers as floats, thus a value the API
// returns as 1.0 against a configuration that writes 1 must not diff.
func TestJSONReadsANumberByValue(t *testing.T) {
	equal, diags := NewJSONValue(`{"temperature":1}`).StringSemanticEquals(
		context.Background(),
		NewJSONValue(`{"temperature":1.0}`),
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !equal {
		t.Fatal("expected 1 and 1.0 to compare equal")
	}
}

func TestJSONIgnoresKeyOrderAndSpacing(t *testing.T) {
	equal, diags := NewJSONValue("{\n  \"a\": 1,\n  \"b\": 2\n}").
		StringSemanticEquals(
			context.Background(),
			NewJSONValue(`{"b":2,"a":1}`),
		)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !equal {
		t.Fatal("expected the two documents to compare equal")
	}
}

func TestJSONEmptyEqualsOnlyEmpty(t *testing.T) {
	equal, _ := NewJSONValue("").StringSemanticEquals(
		context.Background(),
		NewJSONValue(`{"a":1}`),
	)
	if equal {
		t.Fatal("an empty value must not equal a document")
	}
}

// Key order is ignored at every depth, while the order of an array is a
// difference, because an array is ordered.
func TestJSONIgnoresKeyOrderWhenNested(t *testing.T) {
	equal, _ := NewJSONValue(`{"a":{"x":1,"y":[1,2]},"b":2}`).
		StringSemanticEquals(
			context.Background(),
			NewJSONValue(`{"b":2,"a":{"y":[1,2],"x":1}}`),
		)
	if !equal {
		t.Fatal("expected nested keys in another order to compare equal")
	}

	equal, _ = NewJSONValue(`{"a":[1,2]}`).StringSemanticEquals(
		context.Background(),
		NewJSONValue(`{"a":[2,1]}`),
	)
	if equal {
		t.Fatal("expected an array in another order to differ")
	}
}
