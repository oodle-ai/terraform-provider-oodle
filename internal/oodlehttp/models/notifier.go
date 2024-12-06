package models

import (
	"terraform-provider-oodle/internal/oodlehttp/models/oprom"
)

// Notifier represents a single notification target.
type Notifier struct {
	ID              ID                     `json:"id,omitempty" yaml:"id,omitempty"`
	Name            string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Type            NotifierType           `json:"type" yaml:"type"`
	EmailConfig     *oprom.EmailConfig     `json:"email_config,omitempty" yaml:"email_config,omitempty"`
	PagerdutyConfig *oprom.PagerdutyConfig `json:"pagerduty_config,omitempty" yaml:"pagerduty_config,omitempty"`
	SlackConfig     *oprom.SlackConfig     `json:"slack_config,omitempty" yaml:"slack_config,omitempty"`
	OpsGenieConfig  *oprom.OpsGenieConfig  `json:"opsgenie_config,omitempty" yaml:"opsgenie_config,omitempty"`
	WebhookConfig   *oprom.WebhookConfig   `json:"webhook_config,omitempty" yaml:"webhook_config,omitempty"`
}
