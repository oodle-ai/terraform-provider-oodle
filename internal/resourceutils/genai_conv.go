package resourceutils

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Conversion helpers for the GenAI resources.
//
// Several GenAI fields are free-form JSON — filters, variable
// mappings, model params, dataset inputs. They are modelled as
// jsontypes.Normalized, which compares two documents by value, and
// converted here.

// NormalizedToRaw parses a JSON attribute into a json.RawMessage. A null,
// unknown or empty value becomes nil, which the client models drop from the
// request entirely.
func NormalizedToRaw(
	value jsontypes.Normalized,
	attribute string,
) (json.RawMessage, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	text := value.ValueString()
	if text == "" {
		return nil, nil
	}

	if !json.Valid([]byte(text)) {
		return nil, fmt.Errorf("%s is not valid JSON: %s", attribute, text)
	}

	return json.RawMessage(text), nil
}

// RawToNormalized converts JSON from the API into a JSON attribute. The type
// compares two documents by value, thus the key order and the spacing the API
// returns raise no diff, and the prior value is not needed here.
func RawToNormalized(raw json.RawMessage) jsontypes.Normalized {
	if len(raw) == 0 || string(raw) == "null" {
		return jsontypes.NewNormalizedNull()
	}

	return jsontypes.NewNormalizedValue(string(raw))
}

// RawToJSONString converts JSON from the API back into a Terraform string.
//
// It serves an attribute that holds JSON only some of the time, thus it cannot
// be a jsontypes.Normalized: the prompt of a text prompt is plain text. When
// the response is semantically equal to what is already in state, the stored
// text is kept, because the API re-serializes JSON and preserves neither the
// key order nor the spacing.
func RawToJSONString(raw json.RawMessage, prior types.String) types.String {
	if len(raw) == 0 || string(raw) == "null" {
		// An attribute the user never set stays null rather than
		// flipping to "" and showing a diff.
		if prior.IsNull() || prior.IsUnknown() {
			return types.StringNull()
		}

		if priorText := prior.ValueString(); priorText == "" {
			return prior
		}

		return types.StringNull()
	}

	if !prior.IsNull() && !prior.IsUnknown() {
		if jsonEquivalent([]byte(prior.ValueString()), raw) {
			return prior
		}
	}

	return types.StringValue(string(raw))
}

// jsonEquivalent reports whether two JSON documents carry the same
// value, ignoring key order and formatting.
func jsonEquivalent(a, b []byte) bool {
	var left, right any
	if err := json.Unmarshal(a, &left); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &right); err != nil {
		return false
	}

	return reflect.DeepEqual(left, right)
}

// StringListToSlice converts a Terraform list of strings to a Go
// slice. A null or unknown list becomes nil.
func StringListToSlice(
	ctx context.Context,
	list types.List,
) ([]string, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var values []string
	if diags := list.ElementsAs(ctx, &values, false); diags.HasError() {
		return nil, fmt.Errorf("failed to read list: %v", diags.Errors())
	}

	return values, nil
}

// SliceToStringList converts a Go slice to a Terraform list of
// strings. An empty slice becomes a null list when the prior value
// was null, so that an unset optional attribute does not start
// showing a diff against an empty list.
func SliceToStringList(
	ctx context.Context,
	values []string,
	prior types.List,
	diagnosticsOut *diag.Diagnostics,
) types.List {
	if len(values) == 0 && (prior.IsNull() || prior.IsUnknown()) {
		return types.ListNull(types.StringType)
	}

	list, diags := types.ListValueFrom(ctx, types.StringType, values)
	if diags.HasError() {
		diagnosticsOut.Append(diags...)
		return prior
	}

	return list
}
