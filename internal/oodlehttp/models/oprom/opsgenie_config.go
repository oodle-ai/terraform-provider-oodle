package oprom

import (
	"github.com/prometheus/alertmanager/config"
	commonconfig "github.com/prometheus/common/config"
)

// OpsGenieConfig configures notifications via OpsGenie.
// It is copied from prometheus/alertmanager/config with following changes:
// 1. Replaced APIKey type from Secret to string which maintains wire compatibility.
// 2. Removed APIKeyFile.
type OpsGenieConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commonconfig.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIKey       string                           `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	APIURL       *config.URL                      `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	Message      string                           `yaml:"message,omitempty" json:"message,omitempty"`
	Description  string                           `yaml:"description,omitempty" json:"description,omitempty"`
	Source       string                           `yaml:"source,omitempty" json:"source,omitempty"`
	Details      map[string]string                `yaml:"details,omitempty" json:"details,omitempty"`
	Entity       string                           `yaml:"entity,omitempty" json:"entity,omitempty"`
	Responders   []config.OpsGenieConfigResponder `yaml:"responders,omitempty" json:"responders,omitempty"`
	Actions      string                           `yaml:"actions,omitempty" json:"actions,omitempty"`
	Tags         string                           `yaml:"tags,omitempty" json:"tags,omitempty"`
	Note         string                           `yaml:"note,omitempty" json:"note,omitempty"`
	Priority     string                           `yaml:"priority,omitempty" json:"priority,omitempty"`
	UpdateAlerts bool                             `yaml:"update_alerts,omitempty" json:"update_alerts,omitempty"`
}
