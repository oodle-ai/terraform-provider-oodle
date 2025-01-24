package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// WebhookConfig configures notifications via a generic webhook.
type WebhookConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	// URL to send POST request to.
	URL string `yaml:"url" json:"url"`
}
