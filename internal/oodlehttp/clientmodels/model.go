// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clientmodels

type ClientModel interface {
	GetID() string
	//MarshalJSON() ([]byte, error)
	//UnmarshalJSON([]byte) error
}
