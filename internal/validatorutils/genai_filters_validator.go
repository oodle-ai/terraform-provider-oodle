package validatorutils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// genaiFilterNumericTypes are the matcher types the API stores as
// numbers, by the operator words and symbols that mean them.
var genaiFilterNumericTypes = map[string]int{
	"eq": 0, "=": 0, "==": 0,
	"neq": 1, "!=": 1,
	"re": 2, "=~": 2,
	"nre": 3, "!~": 3,
	"oneof": 4, "not_oneof": 5,
}

// genaiFilterTypesText lists the stored matcher types, for messages.
const genaiFilterTypesText = "0 (eq), 1 (neq), 2 (re), 3 (nre), " +
	"4 (oneof), 5 (not_oneof), \"GT\", \"GTE\", \"LT\", \"LTE\""

// genaiFilterMaxNumericType is the highest numbered matcher type.
// Types 4 and 5 match one of a list of values, read from the
// "multi_value" field rather than from "value".
const genaiFilterMaxNumericType = 5

// genaiFilterStringTypes are the matcher types the API stores as
// strings.
var genaiFilterStringTypes = map[string]string{
	"gt": "GT", "gte": "GTE", "lt": "LT", "lte": "LTE",
}

type genaiFiltersValidator struct{}

var _ validator.String = genaiFiltersValidator{}

// NewGenAIFiltersValidator checks the filters of a GenAI evaluator.
//
// The API stores each filter in the form the trace store matches, and
// gives back that form. It refuses a dotted attribute name without a
// "span::" or "resource::" prefix, and it rewrites an operator word
// such as "eq" to its stored type. A configuration that uses a word
// thus never matches what the API gives back: the apply fails with an
// inconsistent result, and each later plan shows a diff. This
// validator refuses both cases at plan time, and tells the user the
// value to write.
func NewGenAIFiltersValidator() validator.String {
	return genaiFiltersValidator{}
}

func (v genaiFiltersValidator) Description(_ context.Context) string {
	return "Each filter names a span::/resource:: attribute or a plain " +
		"column, and has a type of " + genaiFilterTypesText + ". Types 4 " +
		"and 5 take their values in a multi_value list."
}

func (v genaiFiltersValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v genaiFiltersValidator) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	text := req.ConfigValue.ValueString()
	if text == "" {
		return
	}

	// A value that is not a list of objects is not this validator's
	// to refuse: the API stores it as it is.
	var entries []map[string]json.RawMessage
	if json.Unmarshal([]byte(text), &entries) != nil {
		return
	}

	for i, entry := range entries {
		if msg := genaiFilterError(entry); msg != "" {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid evaluator filter",
				fmt.Sprintf("filters[%d]: %s", i, msg),
			)
		}
	}
}

// genaiFilterError reports why one filter would not stay as written,
// or "" when it would.
func genaiFilterError(entry map[string]json.RawMessage) string {
	var name string
	if json.Unmarshal(entry["name"], &name) != nil || name == "" {
		return ""
	}

	if !strings.Contains(name, "::") && strings.Contains(name, ".") {
		return fmt.Sprintf(
			"filter %q names an attribute without its kind: write %q "+
				"for a span attribute or %q for a resource attribute",
			name, "span::"+name, "resource::"+name,
		)
	}

	rawType, ok := entry["type"]
	if !ok {
		return ""
	}

	var number float64
	if json.Unmarshal(rawType, &number) == nil {
		if number != float64(int(number)) || number < 0 ||
			number > genaiFilterMaxNumericType {
			return fmt.Sprintf(
				"filter %q: type %s is not one of %s",
				name, string(rawType), genaiFilterTypesText,
			)
		}

		// A one-of matcher reads its values from multi_value only. Without
		// the list, the filter matches nothing.
		if number >= 4 {
			var values []any
			if json.Unmarshal(entry["multi_value"], &values) != nil ||
				values == nil {
				return fmt.Sprintf(
					"filter %q: type %s needs its values as a "+
						"multi_value list, for example "+
						"\"multi_value\": [\"a\", \"b\"]",
					name, string(rawType),
				)
			}
		}

		return ""
	}

	var word string
	if json.Unmarshal(rawType, &word) != nil {
		return ""
	}

	key := strings.ToLower(strings.TrimSpace(word))
	if want, ok := genaiFilterNumericTypes[key]; ok {
		return fmt.Sprintf(
			"filter %q: write type %q as the number %d; the API stores "+
				"it in that form", name, word, want,
		)
	}

	if want, ok := genaiFilterStringTypes[key]; ok {
		if word == want {
			return ""
		}

		return fmt.Sprintf(
			"filter %q: write type %q as %q; the API stores it in that "+
				"form", name, word, want,
		)
	}

	return fmt.Sprintf(
		"filter %q: type %q is not one of %s",
		name, word, genaiFilterTypesText,
	)
}
