package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// WebhookConfig configures notifications via a generic webhook.
type WebhookConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	// URL to send POST request to.
	URL string `yaml:"url" json:"url"`

	// Payload optionally replaces the default webhook body with a fully custom
	// one. Every string key and value in this (arbitrarily nested) object is
	// rendered as a Go template against the alert notification data.
	Payload map[string]any `yaml:"payload,omitempty" json:"payload,omitempty"`
}
