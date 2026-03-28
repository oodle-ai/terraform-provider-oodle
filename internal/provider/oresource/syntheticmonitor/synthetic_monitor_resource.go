package syntheticmonitor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
	"terraform-provider-oodle/internal/validatorutils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &syntheticMonitorResource{}
	_ resource.ResourceWithConfigure   = &syntheticMonitorResource{}
	_ resource.ResourceWithImportState = &syntheticMonitorResource{}
)

const syntheticMonitorsResourcePath = "synthetic-monitors"

var validRuleTypes = map[string]struct{}{
	"http": {},
}

var validHTTPMethods = map[string]struct{}{
	"GET":     {},
	"POST":    {},
	"PUT":     {},
	"DELETE":  {},
	"PATCH":   {},
	"HEAD":    {},
	"OPTIONS": {},
}

// syntheticMonitorResource is the resource implementation.
type syntheticMonitorResource struct {
	oresource.BaseResource[*clientmodels.SyntheticMonitor, *syntheticMonitorResourceModel]
}

func NewSyntheticMonitorResource() resource.Resource {
	modelCreator := func() *clientmodels.SyntheticMonitor {
		return &clientmodels.SyntheticMonitor{}
	}
	return &syntheticMonitorResource{
		BaseResource: oresource.NewBaseResource[*clientmodels.SyntheticMonitor, *syntheticMonitorResourceModel](
			func() *syntheticMonitorResourceModel {
				return &syntheticMonitorResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[*clientmodels.SyntheticMonitor] {
				return oodlehttp.NewModelClient[*clientmodels.SyntheticMonitor](
					oodleHttpClient,
					syntheticMonitorsResourcePath,
					modelCreator,
				)
			},
		),
	}
}

// Metadata returns the resource type name.
func (r *syntheticMonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_synthetic_monitor"
}

// Schema defines the schema for the resource.
func (r *syntheticMonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a synthetic monitor. Synthetic monitors periodically check the availability and performance of endpoints.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the synthetic monitor.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name for the synthetic monitor.",
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the synthetic monitor is enabled.",
			},
			"rule_type": schema.StringAttribute{
				Required:    true,
				Description: "Type of the synthetic monitor rule. Possible values: 'http'.",
				Validators: []validator.String{
					validatorutils.NewChoiceValidator(validRuleTypes),
				},
			},
			"rule_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Configuration for the synthetic monitor rule.",
				Attributes: map[string]schema.Attribute{
					"http": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "HTTP rule configuration.",
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Required:    true,
								Description: "URL to monitor.",
							},
							"method": schema.StringAttribute{
								Required:    true,
								Description: "HTTP method to use. Possible values: 'GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'.",
								Validators: []validator.String{
									validatorutils.NewChoiceValidator(validHTTPMethods),
								},
							},
							"headers": schema.MapAttribute{
								Optional:    true,
								ElementType: types.StringType,
								Description: "HTTP headers to send with the request.",
							},
							"body": schema.StringAttribute{
								Optional:    true,
								Description: "Request body to send.",
							},
							"expected_status_codes": schema.ListAttribute{
								Required:    true,
								ElementType: types.StringType,
								Description: "List of expected HTTP status codes or patterns (e.g., '200', '2XX').",
							},
							"follow_redirects": schema.BoolAttribute{
								Optional:    true,
								Description: "Whether to follow HTTP redirects.",
							},
							"insecure_skip_verify": schema.BoolAttribute{
								Optional:    true,
								Description: "Whether to skip TLS certificate verification.",
							},
						},
					},
				},
			},
			"interval": schema.StringAttribute{
				Required:    true,
				Description: "Interval between checks (e.g., '30s', '1m').",
				Validators: []validator.String{
					validatorutils.NewDurationValidator(),
				},
			},
			"timeout": schema.StringAttribute{
				Required:    true,
				Description: "Timeout for each check (e.g., '5s', '10s').",
				Validators: []validator.String{
					validatorutils.NewDurationValidator(),
				},
			},
		},
	}
}
