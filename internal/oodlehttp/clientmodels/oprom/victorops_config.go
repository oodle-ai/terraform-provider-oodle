package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// VictorOpsConfig configures notifications via Splunk On-Call (VictorOps).
type VictorOpsConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	// APIKey is the key in the URL of the Splunk On-Call REST endpoint
	// integration. It is a plain string, because the Alertmanager Secret type
	// would send "<secret>" in its place.
	APIKey string `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	// RoutingKey selects the Splunk On-Call escalation policy.
	RoutingKey string `yaml:"routing_key,omitempty" json:"routing_key,omitempty"`
}
