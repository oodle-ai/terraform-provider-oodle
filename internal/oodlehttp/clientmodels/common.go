// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clientmodels

import (
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

// ID is a model for a unique identifier used in all alert related models.
type ID struct {
	UUID uuid.UUID
}

// MarshalJSON customizes the JSON marshaling for ID.
func (id ID) MarshalJSON() ([]byte, error) {
	return jsoniter.Marshal(id.UUID)
}

// UnmarshalJSON customizes the JSON unmarshaling for ID.
func (id *ID) UnmarshalJSON(data []byte) error {
	return jsoniter.Unmarshal(data, &id.UUID)
}
