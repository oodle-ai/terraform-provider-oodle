package oprom

import (
	"github.com/prometheus/alertmanager/config"
	commonconfig "github.com/prometheus/common/config"
)

// WebhookConfig configures notifications via a generic webhook.
// It is copied from prometheus/alertmanager/config with following changes:
// 1. Replaced URL to string which maintains wire compatibility.
// 2. Removed URLFile.
type WebhookConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commonconfig.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	// URL to send POST request to.
	URL string `yaml:"url" json:"url"`

	// MaxAlerts is the maximum number of alerts to be sent per webhook message.
	// Alerts exceeding this threshold will be truncated. Setting this to 0
	// allows an unlimited number of alerts.
	MaxAlerts uint64 `yaml:"max_alerts" json:"max_alerts"`
}
