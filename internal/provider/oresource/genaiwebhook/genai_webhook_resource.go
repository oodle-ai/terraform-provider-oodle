package genaiwebhook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &genaiWebhookResource{}
	_ resource.ResourceWithConfigure   = &genaiWebhookResource{}
	_ resource.ResourceWithImportState = &genaiWebhookResource{}
)

// genaiWebhookResource is the resource implementation.
type genaiWebhookResource struct {
	oresource.APIBaseResource[
		*clientmodels.GenAIWebhook,
		*genaiWebhookResourceModel,
	]
}

func NewGenAIWebhookResource() resource.Resource {
	modelCreator := func() *clientmodels.GenAIWebhook {
		return &clientmodels.GenAIWebhook{}
	}

	return &genaiWebhookResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.GenAIWebhook,
			*genaiWebhookResourceModel,
		](
			func() *genaiWebhookResourceModel {
				return &genaiWebhookResourceModel{}
			},
			modelCreator,
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.GenAIWebhook] {
				return oodlehttp.NewGenAIWebhookClient(oodleHttpClient)
			},
		),
	}
}

func (r *genaiWebhookResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_genai_webhook"
}

// Schema defines the schema for the resource.
func (r *genaiWebhookResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "An experiment webhook: an endpoint you host that " +
			"runs your own agent or workflow. An experiment run against " +
			"it posts every dataset item to the URL, using the request " +
			"template, reads the output out of the reply at the output " +
			"path, and scores it. Every request carries a W3C traceparent " +
			"header, so a service with OpenTelemetry HTTP instrumentation " +
			"links its trace to each result.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the webhook.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the webhook, unique per instance.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "What the endpoint runs.",
			},
			"url": schema.StringAttribute{
				Required: true,
				Description: "URL each item is POSTed to. Must be absolute " +
					"http(s) and reachable from the internet; private and " +
					"cluster-local hosts are refused.",
			},
			"headers": schema.MapAttribute{
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
				Description: "Headers sent on every request, on top of " +
					"Content-Type and traceparent. Put the endpoint's " +
					"credential here. The API stores them encrypted and " +
					"returns only their names, so Terraform cannot detect a " +
					"value changed outside of Terraform.",
			},
			"timeout_seconds": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Description: "How long one item may take, 1 to 600. " +
					"Defaults to 60.",
				Validators: []validator.Int64{
					timeoutRangeValidator{},
				},
			},
			"request_template": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The JSON body sent per item, with {{path}} " +
					"placeholders read from the item ({{input}}, " +
					"{{input.<field>}}, {{metadata.<field>}}, {{id}}) and " +
					"the run ({{run.id}}, {{run.name}}, {{dataset.id}}, " +
					"{{dataset.name}}), inserted as JSON. Empty sends the " +
					"item's input as the body.",
			},
			"output_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Where the output is in the reply, such as " +
					"'answer' or 'choices[0].message.content'. Empty " +
					"stores the whole reply.",
			},
		},
	}
}

// Bounds shared with the API, which refuses a timeout outside them.
const (
	minTimeoutSeconds = 1
	maxTimeoutSeconds = 600
)

// timeoutRangeValidator refuses a timeout the API would refuse, so
// the plan fails rather than the apply.
type timeoutRangeValidator struct{}

func (timeoutRangeValidator) Description(_ context.Context) string {
	return "timeout_seconds must be between 1 and 600"
}

func (v timeoutRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v timeoutRangeValidator) ValidateInt64(
	ctx context.Context,
	req validator.Int64Request,
	resp *validator.Int64Response,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueInt64()
	if value < minTimeoutSeconds || value > maxTimeoutSeconds {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid timeout_seconds",
			v.Description(ctx),
		)
	}
}
