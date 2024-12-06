package oprom

import (
	"github.com/prometheus/alertmanager/config"
	commonconfig "github.com/prometheus/common/config"
)

// PagerdutyConfig configures notifications via PagerDuty.
// It is copied from prometheus/alertmanager/config with following changes:
// 1. Replaced ServiceKey type from Secret to string which maintains wire compatibility.
// 2. Removed ServiceKeyFile.
// 3. Replaced RoutingKey type from Secret to string which maintains wire compatibility.
type PagerdutyConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commonconfig.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	ServiceKey     string                  `yaml:"service_key,omitempty" json:"service_key,omitempty"`
	RoutingKey     string                  `yaml:"routing_key,omitempty" json:"routing_key,omitempty"`
	RoutingKeyFile string                  `yaml:"routing_key_file,omitempty" json:"routing_key_file,omitempty"`
	URL            *config.URL             `yaml:"url,omitempty" json:"url,omitempty"`
	Client         string                  `yaml:"client,omitempty" json:"client,omitempty"`
	ClientURL      string                  `yaml:"client_url,omitempty" json:"client_url,omitempty"`
	Description    string                  `yaml:"description,omitempty" json:"description,omitempty"`
	Details        map[string]string       `yaml:"details,omitempty" json:"details,omitempty"`
	Images         []config.PagerdutyImage `yaml:"images,omitempty" json:"images,omitempty"`
	Links          []config.PagerdutyLink  `yaml:"links,omitempty" json:"links,omitempty"`
	Source         string                  `yaml:"source,omitempty" json:"source,omitempty"`
	Severity       string                  `yaml:"severity,omitempty" json:"severity,omitempty"`
	Class          string                  `yaml:"class,omitempty" json:"class,omitempty"`
	Component      string                  `yaml:"component,omitempty" json:"component,omitempty"`
	Group          string                  `yaml:"group,omitempty" json:"group,omitempty"`
}
