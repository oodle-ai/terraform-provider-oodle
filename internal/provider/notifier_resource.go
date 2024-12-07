package provider

import (
	"context"

	"terraform-provider-oodle/internal/oodlehttp/models"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-oodle/internal/validatorutils"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type notifierResource struct {
	baseResource
}

// NewNotifierResource is a helper function to simplify the provider implementation.
func NewNotifierResource() resource.Resource {
	return &notifierResource{}
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

type notifierModel struct {
	ID              types.String          `tfsdk:"id"`
	Name            types.String          `tfsdk:"name"`
	Type            types.String          `tfsdk:"type"`
	PagerdutyConfig *pagerdutyConfigModel `tfsdk:"pagerduty_config"`
	SlackConfig     *slackConfigModel     `tfsdk:"slack_config"`
	OpsGenieConfig  *opsgenieConfigModel
	WebhookConfig   *webhookConfigModel
}

func (n notifierResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notifier"
}

func (n notifierResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					validatorutils.NewChoiceValidator(models.NotifierNames),
				},
			},
			"pagerduty_config": schema.SingleNestedAttribute{
				Description: "PagerDuty notifier configuration.",
				Attributes: map[string]schema.Attribute{
					"service_key": schema.StringAttribute{
						Required:    true,
						Description: "PagerDuty service key.",
					},
					"send_resolved": schema.BoolAttribute{
						Optional:    true,
						Description: "Send notifications when incident is resolved.",
					},
				},
			},
			"slack_config": schema.SingleNestedAttribute{
				Description: "Slack notifier configuration.",
				Attributes: map[string]schema.Attribute{
					"api_url": schema.StringAttribute{
						Required:    true,
						Description: "Slack API URL.",
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
				Description: "OpsGenie notifier configuration.",
				Attributes: map[string]schema.Attribute{
					"api_key": schema.StringAttribute{
						Required:    true,
						Description: "OpsGenie API key.",
					},
					"send_resolved": schema.BoolAttribute{
						Optional:    true,
						Description: "Send notifications when incident is resolved.",
					},
				},
			},
			"webhook_config": schema.SingleNestedAttribute{
				Description: "Webhook notifier configuration.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Required: true,
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

func (n notifierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (n notifierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (n notifierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (n notifierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}
