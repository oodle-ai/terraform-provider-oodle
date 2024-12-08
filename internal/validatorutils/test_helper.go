// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validatorutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func IsValidForValidator(value types.String, vldtr validator.String) bool {
	var request validator.StringRequest
	request.ConfigValue = value
	var response validator.StringResponse

	vldtr.ValidateString(context.TODO(), request, &response)
	return !response.Diagnostics.HasError()
}
