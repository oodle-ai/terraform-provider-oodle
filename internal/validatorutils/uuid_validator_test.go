// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validatorutils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"
)

func TestNewUUIDValidator(t *testing.T) {
	validator := NewUUIDValidator()
	assert.True(t, IsValidForValidator(types.StringValue("f98878b5-7b9f-409c-bbab-bc2c6a3a5f6d"), validator))
	assert.False(t, IsValidForValidator(types.StringValue("f98878b5-7b9f-409c-bbab-bc6a3a5f6d2"), validator))
	assert.False(t, IsValidForValidator(types.StringValue("foo"), validator))
	assert.True(t, IsValidForValidator(types.StringValue("c3b23877-bbab-4177-bb39-27a839f91bbf"), validator))
}
