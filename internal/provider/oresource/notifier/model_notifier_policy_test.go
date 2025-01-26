package notifier

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/prometheus/alertmanager/config"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels/oprom"
)

func TestNotificationPolicyModel(t *testing.T) {
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
			Type: clientmodels.NotifierConfigWebhook,
			WebhookConfig: &oprom.WebhookConfig{
				URL: "test4",
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
				URL: "https://chat.googleapis.com/v1/spaces/XXXXXX/messages?key=YYYYYY&token=ZZZZZ",
			},
		},
	}

	for _, clientModel := range testCases {
		resourceModel := &notifierResourceModel{}
		diags := &diag.Diagnostics{}
		resourceModel.FromClientModel(clientModel, diags)
		assert.False(t, diags.HasError())

		newClientModel := &clientmodels.Notifier{}
		assert.Nil(t, resourceModel.ToClientModel(newClientModel))

		assert.DeepEqual(t, clientModel, newClientModel)
	}
}
