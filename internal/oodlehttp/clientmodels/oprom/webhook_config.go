package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// WebhookConfig configures notifications via a generic webhook.
// It is copied from prometheus/alertmanager/config with following changes:
// 1. Replaced URL to string which maintains wire compatibility.
// 2. Removed URLFile.
type WebhookConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	// URL to send POST request to.
	URL string `yaml:"url" json:"url"`
}
