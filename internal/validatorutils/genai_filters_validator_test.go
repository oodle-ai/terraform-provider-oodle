package validatorutils

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"
)

func genaiFiltersErrors(value types.String) string {
	req := validator.StringRequest{ConfigValue: value}
	resp := &validator.StringResponse{}
	NewGenAIFiltersValidator().ValidateString(context.Background(), req, resp)

	var details []string
	for _, d := range resp.Diagnostics.Errors() {
		details = append(details, d.Detail())
	}

	return strings.Join(details, "\n")
}

func TestGenAIFiltersValidatorAcceptsStoredForm(t *testing.T) {
	for _, filters := range []string{
		`[]`,
		`[{"name": "span::gen_ai.operation.name", "type": 0, "value": "chat"}]`,
		`[{"name": "resource::service.name", "type": 2, "value": "api-.*"}]`,
		`[{"name": "span::gen_ai.usage.input_tokens", "type": "GT", "value": "100"}]`,
		`[{"name": "service_name", "type": 1, "value": "x"}]`,
		`[{"name": "duration", "type": "LTE", "value": "5"}]`,
		// Entries without a name, and values that are not a list, are
		// stored as they are.
		`[{"type": "eq"}]`,
		`{"name": "gen_ai.system"}`,
	} {
		assert.Equal(
			t, "", genaiFiltersErrors(types.StringValue(filters)),
		)
	}

	assert.Equal(t, "", genaiFiltersErrors(types.StringNull()))
	assert.Equal(t, "", genaiFiltersErrors(types.StringUnknown()))
	assert.Equal(t, "", genaiFiltersErrors(types.StringValue("")))
}

func TestGenAIFiltersValidatorRefusesWords(t *testing.T) {
	cases := map[string]string{
		`[{"name": "span::a", "type": "eq"}]`:   `as the number 0`,
		`[{"name": "span::a", "type": "!="}]`:   `as the number 1`,
		`[{"name": "span::a", "type": "=~"}]`:   `as the number 2`,
		`[{"name": "span::a", "type": "nre"}]`:  `as the number 3`,
		`[{"name": "span::a", "type": "gt"}]`:   `as "GT"`,
		`[{"name": "span::a", "type": "Lte"}]`:  `as "LTE"`,
		`[{"name": "span::a", "type": "like"}]`: `is not one of`,
		`[{"name": "span::a", "type": 4}]`:      `is not one of`,
		`[{"name": "span::a", "type": 1.5}]`:    `is not one of`,
	}
	for filters, want := range cases {
		got := genaiFiltersErrors(types.StringValue(filters))
		assert.True(t, strings.Contains(got, want))
	}
}

func TestGenAIFiltersValidatorRefusesBareDottedName(t *testing.T) {
	got := genaiFiltersErrors(types.StringValue(
		`[{"name": "span::ok", "type": 0}, ` +
			`{"name": "gen_ai.system", "type": 0}]`,
	))
	assert.True(t, strings.Contains(got, "filters[1]"))
	assert.True(t, strings.Contains(got, `"span::gen_ai.system"`))
	assert.True(t, strings.Contains(got, `"resource::gen_ai.system"`))
}
