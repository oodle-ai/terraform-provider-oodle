package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// OpsGenieConfig configures notifications via OpsGenie.
type OpsGenieConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`

	APIKey string `yaml:"api_key,omitempty" json:"api_key,omitempty"`
}
