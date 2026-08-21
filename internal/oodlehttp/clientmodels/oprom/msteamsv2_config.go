package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// MSTeamsV2Config configures notifications via Microsoft Teams workflows.
type MSTeamsV2Config struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	// WebhookURL of the Microsoft Teams workflow to post to.
	WebhookURL string `yaml:"webhook_url,omitempty" json:"webhook_url,omitempty"`
	Title      string `yaml:"title,omitempty" json:"title,omitempty"`
	Text       string `yaml:"text,omitempty" json:"text,omitempty"`
}
