package oprom

import "github.com/prometheus/alertmanager/config"

// EmailConfig configures notifications via email.
type EmailConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	To                    string `yaml:"to,omitempty" json:"to,omitempty"`
}
