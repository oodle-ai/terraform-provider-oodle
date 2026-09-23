package gcpproject

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/validatorutils"
)

func accepts(t *testing.T, v validator.String, value string) bool {
	t.Helper()
	resp := &validator.StringResponse{}
	v.ValidateString(
		context.Background(),
		validator.StringRequest{ConfigValue: types.StringValue(value)},
		resp,
	)
	return !resp.Diagnostics.HasError()
}

// TestProjectIDValidation covers the project attribute. The commonest mistake
// is to give the project number or the display name, both of which the API
// stores and the collector then cannot scrape.
func TestProjectIDValidation(t *testing.T) {
	v := validatorutils.NewRegexValidator(gcpProjectPattern, "project id")

	cases := []struct {
		project string
		valid   bool
	}{
		{"acme-prod", true},
		{"my-project-123", true},
		{"abcdef", true},
		// A project number, not an ID.
		{"123456789012", false},
		// A display name: capitals and spaces are not allowed.
		{"Acme Prod", false},
		// Too short, and a trailing hyphen is rejected by Google too.
		{"acme", false},
		{"acme-prod-", false},
		{"", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.valid, accepts(t, v, c.project), c.project)
	}
}

// TestServiceAccountValidation covers both service-account attributes. A user
// address is accepted by the API and then fails every scrape, because Oodle
// can only impersonate a service account.
func TestServiceAccountValidation(t *testing.T) {
	v := validatorutils.NewRegexValidator(gcpServiceAccountPattern, "service account")

	cases := []struct {
		account string
		valid   bool
	}{
		{"oodle@acme-prod.iam.gserviceaccount.com", true},
		{"gcp-monitoring-integration@oodle-ai.iam.gserviceaccount.com", true},
		// A user address, which cannot be impersonated.
		{"someone@acme.com", false},
		// A Google-managed account outside the .iam. namespace.
		{"acme-prod@appspot.gserviceaccount.com", false},
		{"oodle", false},
		{"", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.valid, accepts(t, v, c.account), c.account)
	}
}
