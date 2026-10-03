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
}

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
		"column, and has a type of 0 (eq), 1 (neq), 2 (re), 3 (nre), " +
		"\"GT\", \"GTE\", \"LT\" or \"LTE\"."
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
		if number != float64(int(number)) || number < 0 || number > 3 {
			return fmt.Sprintf(
				"filter %q: type %s is not one of 0 (eq), 1 (neq), "+
					"2 (re), 3 (nre)", name, string(rawType),
			)
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
		"filter %q: type %q is not one of 0 (eq), 1 (neq), 2 (re), "+
			"3 (nre), \"GT\", \"GTE\", \"LT\", \"LTE\"", name, word,
	)
}
