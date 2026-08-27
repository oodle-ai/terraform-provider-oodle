package notifier

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/prometheus/alertmanager/config"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels/oprom"
)

func TestNotificationPolicyModel(t *testing.T) {
	ctx := context.Background()
	testCases := []*clientmodels.Notifier{
		{
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigPagerduty,
			PagerdutyConfig: &oprom.PagerdutyConfig{
				NotifierConfig: config.NotifierConfig{
					VSendResolved: true,
				},
				ServiceKey: "test2",
			},
		},
		{
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigPagerduty,
			PagerdutyConfig: &oprom.PagerdutyConfig{
				NotifierConfig: config.NotifierConfig{
					VSendResolved: true,
				},
				RoutingKey: "routing-key",
			},
		},
		{
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigOpsGenie,
			OpsGenieConfig: &oprom.OpsGenieConfig{
				NotifierConfig: config.NotifierConfig{
					VSendResolved: true,
				},
				APIKey: "test2",
			},
		},
		{
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigEmail,
			EmailConfig: &oprom.EmailConfig{
				To: "test@example.com",
				NotifierConfig: config.NotifierConfig{
					VSendResolved: true,
				},
			},
		},
		{
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigWebhook,
			WebhookConfig: &oprom.WebhookConfig{
				URL: "test4",
			},
		},
		{
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigWebhook,
			WebhookConfig: &oprom.WebhookConfig{
				NotifierConfig: config.NotifierConfig{
					VSendResolved: true,
				},
				URL: "https://example.com/hooks/oodle",
				Payload: map[string]any{
					"text":   "{{ .CommonLabels.alertname }} is {{ .Status }}",
					"labels": "{{ .CommonLabels | toJson }}",
					"card": map[string]any{
						"sections": []any{
							map[string]any{"header": "{{ .Status }}"},
						},
					},
				},
			},
		},
		{
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigSlack,
			SlackConfig: &oprom.SlackConfig{
				NotifierConfig: config.NotifierConfig{
					VSendResolved: true,
				},
				APIURL:    "test2",
				Channel:   "test3",
				TitleLink: "http://foo.bar",
				Text:      "baz",
			},
		},
		{
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigGoogleChat,
			GoogleChatConfig: &oprom.GoogleChatConfig{
				URL:       "https://chat.googleapis.com/v1/spaces/XXXXXX/messages?key=YYYYYY&token=ZZZZZ",
				Threading: false,
			},
		},
		{
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigMSTeamsV2,
			MSTeamsV2Config: &oprom.MSTeamsV2Config{
				NotifierConfig: config.NotifierConfig{
					VSendResolved: true,
				},
				WebhookURL: "https://example.webhook.office.com/webhookb2/XXXXXX",
				Title:      "alert title",
				Text:       "alert text",
			},
		},
		{
			ID:           clientmodels.ID{UUID: uuid.New()},
			Name:         "test",
			Type:         clientmodels.NotifierConfigRootly,
			RootlyConfig: oprom.NewRootlyConfig(oprom.RootlyWebhookURL, "rootly-bearer-token", true, nil),
		},
		{
			// A Rootly notifier created outside Terraform may carry no bearer
			// token, in which case http_config stays unset on the way back out.
			ID:           clientmodels.ID{UUID: uuid.New()},
			Name:         "test",
			Type:         clientmodels.NotifierConfigRootly,
			RootlyConfig: oprom.NewRootlyConfig(oprom.RootlyWebhookURL, "", false, nil),
		},
		{
			// Rootly is a webhook underneath, so it carries a custom payload
			// the same way.
			ID:   clientmodels.ID{UUID: uuid.New()},
			Name: "test",
			Type: clientmodels.NotifierConfigRootly,
			RootlyConfig: oprom.NewRootlyConfig(
				oprom.RootlyWebhookURL,
				"rootly-bearer-token",
				true,
				map[string]any{
					"text":   "{{ .CommonLabels.alertname }} is {{ .Status }}",
					"labels": "{{ .CommonLabels | toJson }}",
					"card": map[string]any{
						"sections": []any{
							map[string]any{"header": "{{ .Status }}"},
						},
					},
				},
			),
		},
	}

	for _, clientModel := range testCases {
		resourceModel := &notifierResourceModel{}
		diags := &diag.Diagnostics{}
		resourceModel.FromClientModel(ctx, clientModel, diags)
		assert.False(t, diags.HasError())

		newClientModel := &clientmodels.Notifier{}
		assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

		assert.DeepEqual(t, clientModel, newClientModel)
	}
}
