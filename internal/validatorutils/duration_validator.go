// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validatorutils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type durationValidator struct {
}

func NewDurationValidator() validator.String {
	return &durationValidator{}
}

var _ validator.String = (*durationValidator)(nil)

func (d durationValidator) Description(ctx context.Context) string {
	return "Validates that the string is a valid duration"
}

func (d durationValidator) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d durationValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() {
		return
	}

	_, err := time.ParseDuration(request.ConfigValue.ValueString())
	if err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid duration",
			fmt.Sprintf("The value %v is not a valid duration: %v", request.ConfigValue.String(), err))
	}

}

func ShortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}
