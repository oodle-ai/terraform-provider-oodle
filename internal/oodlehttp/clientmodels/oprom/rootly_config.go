package oprom

import (
	"github.com/prometheus/alertmanager/config"
)

// RootlyWebhookURL is the endpoint Rootly Alertmanager alert sources listen on.
// Every Rootly alert source shares this URL; the bearer token is what
// identifies the individual source.
const RootlyWebhookURL = "https://webhooks.rootly.com/webhooks/incoming/alertmanager_webhooks"

const rootlyAuthorizationType = "Bearer"

// RootlyConfig configures notifications to a Rootly Alertmanager alert source.
// On the wire it is a generic webhook config: Oodle stores it under
// `rootly_config` and delivers alerts in Alertmanager webhook format.
type RootlyConfig struct {
	config.NotifierConfig `yaml:",inline" json:",inline"`
	// URL to send POST request to.
	URL string `yaml:"url" json:"url"`
	// MaxAlerts is the maximum number of alerts to be sent per webhook message.
	// Setting this to 0 allows an unlimited number of alerts. It is deliberately
	// not surfaced in the Terraform schema: the Oodle UI hardcodes it to 0 too,
	// so Terraform always sends 0 and will reset a value set out of band.
	MaxAlerts uint64 `yaml:"max_alerts" json:"max_alerts"`
	// HTTPConfig carries the Rootly bearer token.
	HTTPConfig *HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`
}

// NewRootlyConfig builds a Rootly config from the fields the provider exposes.
// An empty bearerToken leaves the Authorization header off entirely, which is
// how a notifier created outside Terraform without a token reads back.
func NewRootlyConfig(url string, bearerToken string, sendResolved bool) *RootlyConfig {
	rootlyConfig := &RootlyConfig{
		NotifierConfig: config.NotifierConfig{VSendResolved: sendResolved},
		URL:            url,
	}

	if bearerToken != "" {
		rootlyConfig.HTTPConfig = &HTTPClientConfig{
			Authorization: &Authorization{
				Type:        rootlyAuthorizationType,
				Credentials: bearerToken,
			},
			FollowRedirects: true,
			EnableHTTP2:     true,
		}
	}

	return rootlyConfig
}

// BearerToken returns the configured Rootly bearer token, or an empty string.
func (r *RootlyConfig) BearerToken() string {
	if r.HTTPConfig == nil || r.HTTPConfig.Authorization == nil {
		return ""
	}

	return r.HTTPConfig.Authorization.Credentials
}
