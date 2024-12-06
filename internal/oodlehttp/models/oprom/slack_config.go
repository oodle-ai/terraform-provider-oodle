package oprom

import (
	"github.com/prometheus/alertmanager/config"
	commonconfig "github.com/prometheus/common/config"
)

// SlackConfig configures notifications via Slack.
// It is copied from prometheus/alertmanager/config with following changes:
// 1. Replaced SecretURL to string which maintains wire compatibility.
// 2. Removed APIURLFile.
type SlackConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commonconfig.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIURL string `yaml:"api_url,omitempty" json:"api_url,omitempty"`

	// Slack channel override, (like #other-channel or @username).
	Channel  string `yaml:"channel,omitempty" json:"channel,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Color    string `yaml:"color,omitempty" json:"color,omitempty"`

	Title       string                `yaml:"title,omitempty" json:"title,omitempty"`
	TitleLink   string                `yaml:"title_link,omitempty" json:"title_link,omitempty"`
	Pretext     string                `yaml:"pretext,omitempty" json:"pretext,omitempty"`
	Text        string                `yaml:"text,omitempty" json:"text,omitempty"`
	Fields      []*config.SlackField  `yaml:"fields,omitempty" json:"fields,omitempty"`
	ShortFields bool                  `yaml:"short_fields" json:"short_fields,omitempty"`
	Footer      string                `yaml:"footer,omitempty" json:"footer,omitempty"`
	Fallback    string                `yaml:"fallback,omitempty" json:"fallback,omitempty"`
	CallbackID  string                `yaml:"callback_id,omitempty" json:"callback_id,omitempty"`
	IconEmoji   string                `yaml:"icon_emoji,omitempty" json:"icon_emoji,omitempty"`
	IconURL     string                `yaml:"icon_url,omitempty" json:"icon_url,omitempty"`
	ImageURL    string                `yaml:"image_url,omitempty" json:"image_url,omitempty"`
	ThumbURL    string                `yaml:"thumb_url,omitempty" json:"thumb_url,omitempty"`
	LinkNames   bool                  `yaml:"link_names" json:"link_names,omitempty"`
	MrkdwnIn    []string              `yaml:"mrkdwn_in,omitempty" json:"mrkdwn_in,omitempty"`
	Actions     []*config.SlackAction `yaml:"actions,omitempty" json:"actions,omitempty"`
}
