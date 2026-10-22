package validatorutils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// aliasValidator refuses an object that sets two attributes that
// are names for the same value. With both set, one of them would
// be dropped without a word.
type aliasValidator struct {
	a, b string
}

var _ validator.Object = aliasValidator{}

// NewAliasValidator returns a validator that refuses an object
// where both a and b are set.
func NewAliasValidator(a, b string) validator.Object {
	return aliasValidator{a: a, b: b}
}

func (v aliasValidator) Description(_ context.Context) string {
	return fmt.Sprintf("%s and %s are the same value; set only one of them", v.a, v.b)
}

func (v aliasValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v aliasValidator) ValidateObject(
	ctx context.Context,
	req validator.ObjectRequest,
	resp *validator.ObjectResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	a, okA := attrs[v.a]
	b, okB := attrs[v.b]
	if okA && okB && !a.IsNull() && !b.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName(v.b),
			"Conflicting attributes",
			v.Description(ctx),
		)
	}
}
