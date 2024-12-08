package validatorutils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"
)

func TestDurationValidator(t *testing.T) {
	validator := NewDurationValidator()

	assert.True(t, IsValidForValidator(types.StringValue("1s"), validator))
	assert.True(t, IsValidForValidator(types.StringValue("1m0s"), validator))
	assert.True(t, IsValidForValidator(types.StringValue("1h15s"), validator))
	assert.True(t, IsValidForValidator(types.StringValue("1m"), validator))
	assert.False(t, IsValidForValidator(types.StringValue("1"), validator))
	assert.False(t, IsValidForValidator(types.StringValue("foo"), validator))

	assert.True(t, IsValidForValidator(types.StringNull(), validator))
}
