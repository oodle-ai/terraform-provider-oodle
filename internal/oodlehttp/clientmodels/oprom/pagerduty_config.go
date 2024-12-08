package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// PagerdutyConfig configures notifications via PagerDuty.
// It is copied from prometheus/alertmanager/config with following changes:
// 1. Replaced ServiceKey type from Secret to string which maintains wire compatibility.
// 2. Removed ServiceKeyFile.
// 3. Replaced RoutingKey type from Secret to string which maintains wire compatibility.
type PagerdutyConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	ServiceKey            string `yaml:"service_key,omitempty" json:"service_key,omitempty"`
}
