// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build go1.10
// +build go1.10

package sdkmath

import "math"

// Round returns the nearest integer, rounding half away from zero.
//
// Special cases are:
//	Round(±0) = ±0
//	Round(±Inf) = ±Inf
//	Round(NaN) = NaN
func Round(x float64) float64 {
	return math.Round(x)
}
