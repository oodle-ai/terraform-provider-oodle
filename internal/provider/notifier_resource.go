package provider

import (
	"context"
	"fmt"

	"terraform-provider-oodle/internal/oodlehttp"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels/oprom"

	"github.com/prometheus/alertmanager/config"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-oodle/internal/validatorutils"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const notifiersResource = "notifiers"

type notifierResource struct {
	baseResource[*clientmodels.Notifier, *notifierResourceModel]
}

func NewNotifierResource() resource.Resource {
	modelCreator := func() *clientmodels.Notifier {
		return &clientmodels.Notifier{}
	}

	return &notifierResource{
		baseResource: newBaseResource(
			func() *notifierResourceModel {
				return &notifierResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[*clientmodels.Notifier] {
				return oodlehttp.NewModelClient[*clientmodels.Notifier](
					oodleHttpClient,
					notifiersResource,
					modelCreator,
				)
			},
		),
	}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &notifierResource{}
	_ resource.ResourceWithConfigure   = &notifierResource{}
	_ resource.ResourceWithImportState = &notifierResource{}
)

type notifierConfigCommonModel struct {
	SendResolved types.Bool `tfsdk:"send_resolved"`
}

type pagerdutyConfigModel struct {
	notifierConfigCommonModel
	ServiceKey types.String `tfsdk:"service_key"`
}

type slackConfigModel struct {
	notifierConfigCommonModel
	APIURL    types.String `tfsdk:"api_url"`
	Channel   types.String `tfsdk:"channel"`
	TitleLink types.String `tfsdk:"title_link"`
	Text      types.String `tfsdk:"text"`
}

type opsgenieConfigModel struct {
	notifierConfigCommonModel
	APIKey types.String `tfsdk:"api_key"`
}

type webhookConfigModel struct {
	notifierConfigCommonModel
	URL types.String `tfsdk:"url"`
}

type notifierResourceModel struct {
	ID              types.String          `tfsdk:"id"`
	Name            types.String          `tfsdk:"name"`
	Type            types.String          `tfsdk:"type"`
	PagerdutyConfig *pagerdutyConfigModel `tfsdk:"pagerduty_config"`
	SlackConfig     *slackConfigModel     `tfsdk:"slack_config"`
	OpsGenieConfig  *opsgenieConfigModel  `tfsdk:"opsgenie_config"`
	WebhookConfig   *webhookConfigModel   `tfsdk:"webhook_config"`
}

func (n *notifierResourceModel) GetID() types.String {
	return n.ID
}

func (n *notifierResourceModel) SetID(id types.String) {
	n.ID = id
}

func (n *notifierResourceModel) FromModel(
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
	case clientmodels.NotifierConfigPagerduty:
		if model.PagerdutyConfig == nil {
			diagnosticsOut.AddError("Missing PagerDuty config", "PagerDuty config is required for PagerDuty notifier")
			return
		}

		n.PagerdutyConfig = &pagerdutyConfigModel{}
		n.PagerdutyConfig.SendResolved = types.BoolValue(model.PagerdutyConfig.SendResolved())
		n.PagerdutyConfig.ServiceKey = types.StringValue(model.PagerdutyConfig.ServiceKey)
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
	default:
		diagnosticsOut.AddError("Unknown type", fmt.Sprintf("Unknown notifier type %v", model.Type))
		return
	}
}

func (n *notifierResourceModel) ToModel(
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
	case clientmodels.NotifierConfigPagerduty:
		if n.PagerdutyConfig == nil {
			return fmt.Errorf("missing PagerDuty config")
		}

		model.PagerdutyConfig = &oprom.PagerdutyConfig{
			ServiceKey: n.PagerdutyConfig.ServiceKey.ValueString(),
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
	default:
		return fmt.Errorf("unknown notifier type %v", model.Type)
	}

	return nil
}

func (n *notifierResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notifier"
}

func (n *notifierResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the monitor.",
				Validators: []validator.String{
					validatorutils.NewUUIDValidator(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the notifier.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the notifier.",
				Required:    true,
				Validators: []validator.String{
					validatorutils.NewChoiceValidator(clientmodels.NotifierNames),
				},
			},
			"pagerduty_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "PagerDuty notifier configuration.",
				Attributes: map[string]schema.Attribute{
					"service_key": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "PagerDuty service key.",
					},
					"send_resolved": schema.BoolAttribute{
						Optional:    true,
						Description: "Send notifications when incident is resolved.",
					},
				},
			},
			"slack_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Slack notifier configuration.",
				Attributes: map[string]schema.Attribute{
					"api_url": schema.StringAttribute{
						Required:    true,
						Description: "Slack API URL.",
						Sensitive:   true,
					},
					"channel": schema.StringAttribute{
						Required:    true,
						Description: "Slack channel to post notifications in.",
					},
					"title_link": schema.StringAttribute{
						Optional:    true,
						Description: "Optional link to include in the notification title.",
					},
					"text": schema.StringAttribute{
						Required:    true,
						Description: "Additional text to add to the notification.",
					},
					"send_resolved": schema.BoolAttribute{
						Optional:    true,
						Description: "Send notifications when incident is resolved.",
					},
				},
			},
			"opsgenie_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "OpsGenie notifier configuration.",
				Attributes: map[string]schema.Attribute{
					"api_key": schema.StringAttribute{
						Required:    true,
						Description: "OpsGenie API key.",
						Sensitive:   true,
					},
					"send_resolved": schema.BoolAttribute{
						Optional:    true,
						Description: "Send notifications when incident is resolved.",
					},
				},
			},
			"webhook_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Webhook notifier configuration.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
					"send_resolved": schema.BoolAttribute{
						Optional:    true,
						Description: "Send notifications when incident is resolved.",
					},
				},
			},
		},
	}
}
