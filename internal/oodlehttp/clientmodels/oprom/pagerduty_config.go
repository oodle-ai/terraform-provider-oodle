package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// PagerdutyConfig configures notifications via PagerDuty.
type PagerdutyConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	ServiceKey            string `yaml:"service_key,omitempty" json:"service_key,omitempty"`
	RoutingKey            string `yaml:"routing_key,omitempty" json:"routing_key,omitempty"`
}
