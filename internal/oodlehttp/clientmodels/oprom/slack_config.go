// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// SlackConfig configures notifications via Slack.
// It is copied from prometheus/alertmanager/config with following changes:
// 1. Replaced SecretURL to string which maintains wire compatibility.
// 2. Removed APIURLFile.
type SlackConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`

	APIURL    string `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	Channel   string `yaml:"channel,omitempty" json:"channel,omitempty"`
	TitleLink string `yaml:"title_link,omitempty" json:"title_link,omitempty"`
	Text      string `yaml:"text,omitempty" json:"text,omitempty"`
}
