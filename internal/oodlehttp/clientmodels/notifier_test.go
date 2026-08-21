package clientmodels

import (
	"encoding/json"
	"testing"

	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels/oprom"
)

// TestRootlyNotifierWireFormat pins the JSON the provider sends for a Rootly
// notifier to what the Oodle API expects. The round-trip test in the notifier
// resource package stays in memory and never marshals, so this is the only
// guard against a wrong JSON key, a wrong type integer, or a bearer token
// redacted to "<secret>" - all of which fail silently at runtime.
func TestRootlyNotifierWireFormat(t *testing.T) {
	notifier := &Notifier{
		Name:         "rootly",
		Type:         NotifierConfigRootly,
		RootlyConfig: oprom.NewRootlyConfig(oprom.RootlyWebhookURL, "bearer-token", true),
	}

	encoded, err := json.Marshal(notifier)
	assert.Nil(t, err)

	var decoded map[string]any
	assert.Nil(t, json.Unmarshal(encoded, &decoded))

	assert.Equal(t, float64(7), decoded["type"])
	assert.DeepEqual(t, map[string]any{
		"url":           oprom.RootlyWebhookURL,
		"max_alerts":    float64(0),
		"send_resolved": true,
		"http_config": map[string]any{
			"authorization": map[string]any{
				"type":        "Bearer",
				"credentials": "bearer-token",
			},
			"follow_redirects": true,
			"enable_http2":     true,
		},
	}, decoded["rootly_config"])
}

func TestMSTeamsV2NotifierWireFormat(t *testing.T) {
	notifier := &Notifier{
		Name: "msteams",
		Type: NotifierConfigMSTeamsV2,
		MSTeamsV2Config: &oprom.MSTeamsV2Config{
			WebhookURL: "https://example.webhook.office.com/webhookb2/XXXXXX",
		},
	}

	encoded, err := json.Marshal(notifier)
	assert.Nil(t, err)

	var decoded map[string]any
	assert.Nil(t, json.Unmarshal(encoded, &decoded))

	assert.Equal(t, float64(6), decoded["type"])
	assert.DeepEqual(t, map[string]any{
		"webhook_url":   "https://example.webhook.office.com/webhookb2/XXXXXX",
		"send_resolved": false,
	}, decoded["msteamsv2_config"])
}
