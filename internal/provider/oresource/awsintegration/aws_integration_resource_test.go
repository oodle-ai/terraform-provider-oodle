package awsintegration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/validatorutils"
)

// TestRoleArnNamePrefixValidation covers the role_arn name-prefix guard: Oodle
// can only assume roles named OodleIntegration*, so a differently-named role
// must be rejected at plan time rather than failing with an opaque
// "Failed to assume IAM role" at apply time.
func TestRoleArnNamePrefixValidation(t *testing.T) {
	ctx := context.Background()
	v := validatorutils.NewRegexValidator(awsRoleNamePattern, "role name prefix")

	accepts := func(arn string) bool {
		resp := &validator.StringResponse{}
		v.ValidateString(ctx, validator.StringRequest{ConfigValue: types.StringValue(arn)}, resp)
		return !resp.Diagnostics.HasError()
	}

	cases := []struct {
		arn   string
		valid bool
	}{
		{"arn:aws:iam::123456789012:role/OodleIntegrationRole", true},
		{"arn:aws:iam::123456789012:role/OodleIntegrationRole-tf", true},
		{"arn:aws:iam::123456789012:role/OodleIntegrationCloudWatchMetricsTagBasedDiscoveryRole", true},
		// Wrong prefix: "OodleAWS..." is the exact mistake that reached apply and
		// failed the assume-role.
		{"arn:aws:iam::123456789012:role/OodleAWSIntegrationRole", false},
		{"arn:aws:iam::123456789012:role/OodleAWSIntegrationRole-tf", false},
		{"arn:aws:iam::123456789012:role/SomeOtherRole", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.valid, accepts(c.arn), c.arn)
	}
}
