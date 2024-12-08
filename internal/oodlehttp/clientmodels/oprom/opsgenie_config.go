// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// OpsGenieConfig configures notifications via OpsGenie.
// It is copied from prometheus/alertmanager/config with following changes:
// 1. Replaced APIKey type from Secret to string which maintains wire compatibility.
// 2. Removed APIKeyFile.
type OpsGenieConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`

	APIKey string `yaml:"api_key,omitempty" json:"api_key,omitempty"`
}
