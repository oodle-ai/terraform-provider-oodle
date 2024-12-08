// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validatorutils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"
)

func TestPromqlValidator(t *testing.T) {
	validator := NewPromQLValidator()
	assert.True(t, IsValidForValidator(types.StringValue("sum(rate(foo[5m]))"), validator))
	assert.True(t, IsValidForValidator(types.StringValue("foo"), validator))
	assert.False(t, IsValidForValidator(types.StringValue("foo[)"), validator))
	assert.False(t, IsValidForValidator(types.StringValue("unknown(test)"), validator))

	assert.True(t, IsValidForValidator(types.StringNull(), validator))
}
