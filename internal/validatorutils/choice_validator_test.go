package validatorutils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"
)

func TestChoiceValidator(t *testing.T) {
	validator := NewChoiceValidator(map[string]struct{}{
		"foo": {},
		"bar": {},
		"baz": {},
	})

	assert.True(t, IsValidForValidator(types.StringValue("foo"), validator))
	assert.False(t, IsValidForValidator(types.StringValue("foo2"), validator))
	assert.True(t, IsValidForValidator(types.StringValue("bar"), validator))
	assert.False(t, IsValidForValidator(types.StringValue("test"), validator))
	assert.True(t, IsValidForValidator(types.StringValue("baz"), validator))

	assert.True(t, IsValidForValidator(types.StringNull(), validator))
}
