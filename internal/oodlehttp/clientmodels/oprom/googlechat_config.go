package oprom

import "github.com/prometheus/alertmanager/config"

// GoogleChatConfig configures notifications via Google Chat.
type GoogleChatConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	URL                   string `yaml:"url,omitempty" json:"url,omitempty"`
}
