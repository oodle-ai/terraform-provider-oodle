// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clientmodels

import (
	"terraform-provider-oodle/internal/oodlehttp/clientmodels/oprom"
)

// Notifier represents a single notification target.
type Notifier struct {
	ID              ID                     `json:"id,omitempty" yaml:"id,omitempty"`
	Name            string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Type            NotifierType           `json:"type" yaml:"type"`
	PagerdutyConfig *oprom.PagerdutyConfig `json:"pagerduty_config,omitempty" yaml:"pagerduty_config,omitempty"`
	SlackConfig     *oprom.SlackConfig     `json:"slack_config,omitempty" yaml:"slack_config,omitempty"`
	OpsGenieConfig  *oprom.OpsGenieConfig  `json:"opsgenie_config,omitempty" yaml:"opsgenie_config,omitempty"`
	WebhookConfig   *oprom.WebhookConfig   `json:"webhook_config,omitempty" yaml:"webhook_config,omitempty"`
}

func (n *Notifier) GetID() string {
	return n.ID.UUID.String()
}
