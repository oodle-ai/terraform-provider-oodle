package notifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
	"terraform-provider-oodle/internal/validatorutils"
)

const notifiersResource = "notifiers"

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &notifierResource{}
	_ resource.ResourceWithConfigure   = &notifierResource{}
	_ resource.ResourceWithImportState = &notifierResource{}
)

type notifierResource struct {
	oresource.BaseResource[*clientmodels.Notifier, *notifierResourceModel]
}

func NewNotifierResource() resource.Resource {
	modelCreator := func() *clientmodels.Notifier {
		return &clientmodels.Notifier{}
	}

	return &notifierResource{
		BaseResource: oresource.NewBaseResource(
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

func (n *notifierResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notifier"
}

func (n *notifierResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the notifier.",
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
						Computed:    true,
						Default:     booldefault.StaticBool(false),
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
						Computed:    true,
						Default:     validatorutils.NewDefaultString(types.StringValue("")),
						Description: "Link to be included in the notification title.",
					},
					"text": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     validatorutils.NewDefaultString(types.StringValue("")),
						Description: "Text to be included in the Slack notification.",
					},
					"send_resolved": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
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
						Computed:    true,
						Default:     booldefault.StaticBool(false),
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
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Send notifications when incident is resolved.",
					},
				},
			},
			"googlechat_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Google chat notifier configuration.",
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
					"send_resolved": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Send notifications when incident is resolved.",
					},
				},
			},
		},
	}
}
