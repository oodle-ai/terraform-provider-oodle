package oprom

import (
	"github.com/prometheus/alertmanager/config"
	commonconfig "github.com/prometheus/common/config"
)

// EmailConfig configures notifications via mail.
type EmailConfig struct {
	// It is copied from prometheus/alertmanager/config with following changes:
	// 1. Replaced AuthPassword type from Secret to string which maintains wire compatibility.
	// 2. Removed AuthPasswordFile.
	// 3. Replaced AuthSecret type from Secret to string which maintains wire compatibility.
	config.NotifierConfig `yaml:",inline" json:",inline"`

	// Email address to notify.
	To           string                 `yaml:"to,omitempty" json:"to,omitempty"`
	From         string                 `yaml:"from,omitempty" json:"from,omitempty"`
	Hello        string                 `yaml:"hello,omitempty" json:"hello,omitempty"`
	Smarthost    config.HostPort        `yaml:"smarthost,omitempty" json:"smarthost,omitempty"`
	AuthUsername string                 `yaml:"auth_username,omitempty" json:"auth_username,omitempty"`
	AuthPassword string                 `yaml:"auth_password,omitempty" json:"auth_password,omitempty"`
	AuthSecret   string                 `yaml:"auth_secret,omitempty" json:"auth_secret,omitempty"`
	AuthIdentity string                 `yaml:"auth_identity,omitempty" json:"auth_identity,omitempty"`
	Headers      map[string]string      `yaml:"headers,omitempty" json:"headers,omitempty"`
	HTML         string                 `yaml:"html,omitempty" json:"html,omitempty"`
	Text         string                 `yaml:"text,omitempty" json:"text,omitempty"`
	RequireTLS   *bool                  `yaml:"require_tls,omitempty" json:"require_tls,omitempty"`
	TLSConfig    commonconfig.TLSConfig `yaml:"tls_config,omitempty" json:"tls_config,omitempty"`
}
