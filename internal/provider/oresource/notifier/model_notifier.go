package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prometheus/alertmanager/config"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels/oprom"
	"terraform-provider-oodle/internal/resourceutils"
)

type notifierResourceModel struct {
	ID               types.String           `tfsdk:"id"`
	Name             types.String           `tfsdk:"name"`
	Type             types.String           `tfsdk:"type"`
	EmailConfig      *emailConfigModel      `tfsdk:"email_config"`
	PagerdutyConfig  *pagerdutyConfigModel  `tfsdk:"pagerduty_config"`
	SlackConfig      *slackConfigModel      `tfsdk:"slack_config"`
	OpsGenieConfig   *opsgenieConfigModel   `tfsdk:"opsgenie_config"`
	WebhookConfig    *webhookConfigModel    `tfsdk:"webhook_config"`
	GoogleChatConfig *googleChatConfigModel `tfsdk:"googlechat_config"`
	MSTeamsV2Config  *msTeamsV2ConfigModel  `tfsdk:"msteamsv2_config"`
	RootlyConfig     *rootlyConfigModel     `tfsdk:"rootly_config"`
}

var _ resourceutils.ResourceModel[*clientmodels.Notifier] = (*notifierResourceModel)(nil)

func (n *notifierResourceModel) GetID() types.String {
	return n.ID
}

func (n *notifierResourceModel) SetID(id types.String) {
	n.ID = id
}

func (n *notifierResourceModel) FromClientModel(
	ctx context.Context,
	model *clientmodels.Notifier,
	diagnosticsOut *diag.Diagnostics,
) {
	n.ID = types.StringValue(model.ID.UUID.String())
	n.Name = types.StringValue(model.Name)
	n.Type = types.StringValue(model.Type.String())
	if len(n.Type.ValueString()) == 0 {
		diagnosticsOut.AddError("Unknown type", fmt.Sprintf("Unknown notifier type %v", model.Type))
	}

	switch model.Type {
	case clientmodels.NotifierConfigEmail:
		if model.EmailConfig == nil {
			diagnosticsOut.AddError("Missing email config", "Email config is required for email notifier")
			return
		}

		n.EmailConfig = &emailConfigModel{}
		n.EmailConfig.SendResolved = types.BoolValue(model.EmailConfig.SendResolved())
		n.EmailConfig.To = types.StringValue(model.EmailConfig.To)
	case clientmodels.NotifierConfigPagerduty:
		if model.PagerdutyConfig == nil {
			diagnosticsOut.AddError("Missing PagerDuty config", "PagerDuty config is required for PagerDuty notifier")
			return
		}

		n.PagerdutyConfig = &pagerdutyConfigModel{}
		n.PagerdutyConfig.SendResolved = types.BoolValue(model.PagerdutyConfig.SendResolved())
		n.PagerdutyConfig.ServiceKey = types.StringValue(model.PagerdutyConfig.ServiceKey)
		n.PagerdutyConfig.RoutingKey = types.StringValue(model.PagerdutyConfig.RoutingKey)
	case clientmodels.NotifierConfigSlack:
		if model.SlackConfig == nil {
			diagnosticsOut.AddError("Missing Slack config", "Slack config is required for Slack notifier")
			return
		}

		n.SlackConfig = &slackConfigModel{}
		n.SlackConfig.SendResolved = types.BoolValue(model.SlackConfig.SendResolved())
		n.SlackConfig.APIURL = types.StringValue(model.SlackConfig.APIURL)
		n.SlackConfig.Channel = types.StringValue(model.SlackConfig.Channel)
		n.SlackConfig.TitleLink = types.StringValue(model.SlackConfig.TitleLink)
		n.SlackConfig.Text = types.StringValue(model.SlackConfig.Text)
	case clientmodels.NotifierConfigOpsGenie:
		if model.OpsGenieConfig == nil {
			diagnosticsOut.AddError("Missing OpsGenie config", "OpsGenie config is required for OpsGenie notifier")
			return
		}

		n.OpsGenieConfig = &opsgenieConfigModel{}
		n.OpsGenieConfig.SendResolved = types.BoolValue(model.OpsGenieConfig.SendResolved())
		n.OpsGenieConfig.APIKey = types.StringValue(model.OpsGenieConfig.APIKey)
	case clientmodels.NotifierConfigWebhook:
		if model.WebhookConfig == nil {
			diagnosticsOut.AddError("Missing Webhook config", "Webhook config is required for Webhook notifier")
			return
		}

		n.WebhookConfig = &webhookConfigModel{}
		n.WebhookConfig.SendResolved = types.BoolValue(model.WebhookConfig.SendResolved())
		n.WebhookConfig.URL = types.StringValue(model.WebhookConfig.URL)
		n.WebhookConfig.Payload = encodePayload(model.WebhookConfig.Payload, "Webhook", diagnosticsOut)
		if diagnosticsOut.HasError() {
			return
		}
	case clientmodels.NotifierConfigGoogleChat:
		if model.GoogleChatConfig == nil {
			diagnosticsOut.AddError("Missing Google chat config", "Google chat config is required for Google chat notifier")
			return
		}

		n.GoogleChatConfig = &googleChatConfigModel{}
		n.GoogleChatConfig.SendResolved = types.BoolValue(model.GoogleChatConfig.SendResolved())
		n.GoogleChatConfig.URL = types.StringValue(model.GoogleChatConfig.URL)
		n.GoogleChatConfig.Threading = types.BoolValue(model.GoogleChatConfig.Threading)
	case clientmodels.NotifierConfigMSTeamsV2:
		if model.MSTeamsV2Config == nil {
			diagnosticsOut.AddError("Missing Microsoft Teams config", "Microsoft Teams config is required for Microsoft Teams notifier")
			return
		}

		n.MSTeamsV2Config = &msTeamsV2ConfigModel{}
		n.MSTeamsV2Config.SendResolved = types.BoolValue(model.MSTeamsV2Config.SendResolved())
		n.MSTeamsV2Config.WebhookURL = types.StringValue(model.MSTeamsV2Config.WebhookURL)
		n.MSTeamsV2Config.Title = types.StringValue(model.MSTeamsV2Config.Title)
		n.MSTeamsV2Config.Text = types.StringValue(model.MSTeamsV2Config.Text)
	case clientmodels.NotifierConfigRootly:
		if model.RootlyConfig == nil {
			diagnosticsOut.AddError("Missing Rootly config", "Rootly config is required for Rootly notifier")
			return
		}

		n.RootlyConfig = &rootlyConfigModel{}
		n.RootlyConfig.SendResolved = types.BoolValue(model.RootlyConfig.SendResolved())
		n.RootlyConfig.URL = types.StringValue(model.RootlyConfig.URL)
		n.RootlyConfig.BearerToken = types.StringValue(model.RootlyConfig.BearerToken())
		n.RootlyConfig.Payload = encodePayload(model.RootlyConfig.Payload, "Rootly", diagnosticsOut)
		if diagnosticsOut.HasError() {
			return
		}
	default:
		diagnosticsOut.AddError("Unknown type", fmt.Sprintf("Unknown notifier type %v", model.Type))
		return
	}
}

func (n *notifierResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.Notifier,
) error {
	var err error
	if !n.ID.IsNull() && !n.ID.IsUnknown() {
		model.ID.UUID, err = uuid.Parse(n.ID.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse ID UUID %v: %v", n.ID.ValueString(), err)
		}
	}

	model.Name = n.Name.ValueString()
	model.Type, err = clientmodels.GetNotifierType(n.Type.ValueString())
	if err != nil {
		return fmt.Errorf("failed to parse notifier type %v: %v", n.Type.ValueString(), err)
	}

	switch model.Type {
	case clientmodels.NotifierConfigEmail:
		if n.EmailConfig == nil {
			return fmt.Errorf("missing email config")
		}

		model.EmailConfig = &oprom.EmailConfig{
			To: n.EmailConfig.To.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.EmailConfig.SendResolved.ValueBool(),
			},
		}
	case clientmodels.NotifierConfigPagerduty:
		if n.PagerdutyConfig == nil {
			return fmt.Errorf("missing PagerDuty config")
		}

		model.PagerdutyConfig = &oprom.PagerdutyConfig{
			ServiceKey: n.PagerdutyConfig.ServiceKey.ValueString(),
			RoutingKey: n.PagerdutyConfig.RoutingKey.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.PagerdutyConfig.SendResolved.ValueBool(),
			},
		}
	case clientmodels.NotifierConfigSlack:
		if n.SlackConfig == nil {
			return fmt.Errorf("missing Slack config")
		}

		model.SlackConfig = &oprom.SlackConfig{
			APIURL:    n.SlackConfig.APIURL.ValueString(),
			Channel:   n.SlackConfig.Channel.ValueString(),
			TitleLink: n.SlackConfig.TitleLink.ValueString(),
			Text:      n.SlackConfig.Text.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.SlackConfig.SendResolved.ValueBool(),
			},
		}
	case clientmodels.NotifierConfigOpsGenie:
		if n.OpsGenieConfig == nil {
			return fmt.Errorf("missing OpsGenie config")
		}

		model.OpsGenieConfig = &oprom.OpsGenieConfig{
			APIKey: n.OpsGenieConfig.APIKey.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.OpsGenieConfig.SendResolved.ValueBool(),
			},
		}
	case clientmodels.NotifierConfigWebhook:
		if n.WebhookConfig == nil {
			return fmt.Errorf("missing Webhook config")
		}

		model.WebhookConfig = &oprom.WebhookConfig{
			URL: n.WebhookConfig.URL.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.WebhookConfig.SendResolved.ValueBool(),
			},
		}

		payload, err := decodePayload(n.WebhookConfig.Payload)
		if err != nil {
			return fmt.Errorf("webhook config: %v", err)
		}
		model.WebhookConfig.Payload = payload
	case clientmodels.NotifierConfigGoogleChat:
		if n.GoogleChatConfig == nil {
			return fmt.Errorf("missing Google chat config")
		}

		model.GoogleChatConfig = &oprom.GoogleChatConfig{
			URL:       n.GoogleChatConfig.URL.ValueString(),
			Threading: n.GoogleChatConfig.Threading.ValueBool(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.GoogleChatConfig.SendResolved.ValueBool(),
			},
		}
	case clientmodels.NotifierConfigMSTeamsV2:
		if n.MSTeamsV2Config == nil {
			return fmt.Errorf("missing Microsoft Teams config")
		}

		model.MSTeamsV2Config = &oprom.MSTeamsV2Config{
			WebhookURL: n.MSTeamsV2Config.WebhookURL.ValueString(),
			Title:      n.MSTeamsV2Config.Title.ValueString(),
			Text:       n.MSTeamsV2Config.Text.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.MSTeamsV2Config.SendResolved.ValueBool(),
			},
		}
	case clientmodels.NotifierConfigRootly:
		if n.RootlyConfig == nil {
			return fmt.Errorf("missing Rootly config")
		}

		rootlyPayload, err := decodePayload(n.RootlyConfig.Payload)
		if err != nil {
			return fmt.Errorf("rootly config: %v", err)
		}

		model.RootlyConfig = oprom.NewRootlyConfig(
			n.RootlyConfig.URL.ValueString(),
			n.RootlyConfig.BearerToken.ValueString(),
			n.RootlyConfig.SendResolved.ValueBool(),
			rootlyPayload,
		)
	default:
		return fmt.Errorf("unknown notifier type %v", model.Type)
	}

	return nil
}

// encodePayload renders a notifier's custom payload for Terraform state. An
// empty payload is null rather than "{}", so a notifier without one does not
// show up as a diff against a configuration that omits the attribute.
func encodePayload(
	payload map[string]any,
	notifier string,
	diagnosticsOut *diag.Diagnostics,
) resourceutils.JSON {
	if len(payload) == 0 {
		return resourceutils.NewJSONNull()
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		diagnosticsOut.AddError(
			fmt.Sprintf("Invalid %s payload", notifier),
			fmt.Sprintf("Failed to encode the custom payload: %v", err),
		)
		return resourceutils.NewJSONNull()
	}

	return resourceutils.NewJSONValue(string(encoded))
}

// decodePayload parses a custom payload out of Terraform state. Unset, unknown
// and blank all mean "no custom payload", which leaves the notifier sending
// the default Alertmanager body.
func decodePayload(value resourceutils.JSON) (map[string]any, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	raw := value.ValueString()
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	payload := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("failed to parse payload as a JSON object: %v", err)
	}

	return payload, nil
}
