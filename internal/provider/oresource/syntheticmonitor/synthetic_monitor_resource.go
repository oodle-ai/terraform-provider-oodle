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
	"http":      {},
	"multistep": {},
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

var validExtractSources = map[string]struct{}{
	"body":   {},
	"header": {},
}

var validExtractParsers = map[string]struct{}{
	"jsonpath":     {},
	"regex":        {},
	"header_value": {},
}

// httpConfigAttributes returns the schema attributes for an HTTP request
// configuration. It is shared by the single-step "http" rule config and by
// each step's request in a multi-step monitor. expectedStatusCodesRequired
// controls whether expected_status_codes is required (single-step) or optional
// (multi-step steps, where the server defaults to any 2XX).
func httpConfigAttributes(expectedStatusCodesRequired bool) map[string]schema.Attribute {
	expectedStatusCodes := schema.ListAttribute{
		ElementType: types.StringType,
		Description: "List of expected HTTP status codes or patterns (e.g., '200', '2XX').",
	}
	if expectedStatusCodesRequired {
		expectedStatusCodes.Required = true
	} else {
		expectedStatusCodes.Optional = true
	}

	return map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Required:    true,
			Description: "URL to monitor. In multi-step monitors this may reference variables extracted from earlier steps using '{{VAR_NAME}}' syntax.",
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
		"expected_status_codes": expectedStatusCodes,
		"excluded_status_codes": schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "List of status codes or patterns that cause the check to fail (e.g., '500', '5XX').",
		},
		"expected_body": schema.StringAttribute{
			Optional:    true,
			Description: "Substring that must appear in the response body.",
		},
		"max_response_time_ms": schema.Int64Attribute{
			Optional:    true,
			Description: "Fail the check if the response takes longer than this many milliseconds.",
		},
		"expected_headers": schema.MapAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: "Response headers that must match the given values.",
		},
		"follow_redirects": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether to follow HTTP redirects. Defaults to false.",
		},
		"insecure_skip_verify": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether to skip TLS certificate verification. Defaults to false.",
		},
		"basic_auth": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "HTTP basic authentication credentials.",
			Attributes: map[string]schema.Attribute{
				"username": schema.StringAttribute{
					Required:    true,
					Description: "Basic auth username.",
				},
				"password": schema.StringAttribute{
					Required:    true,
					Sensitive:   true,
					Description: "Basic auth password.",
				},
			},
		},
		"bearer_token": schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "Bearer token sent as an 'Authorization: Bearer <token>' header.",
		},
	}
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
				Description: "Type of the synthetic monitor rule. Possible values: 'http', 'multistep'.",
				Validators: []validator.String{
					validatorutils.NewChoiceValidator(validRuleTypes),
				},
			},
			"rule_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Configuration for the synthetic monitor rule. Set 'http' for a single-step monitor (rule_type 'http') or 'multistep' for a multi-step monitor (rule_type 'multistep').",
				Attributes: map[string]schema.Attribute{
					"http": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "HTTP rule configuration. Used when rule_type is 'http'.",
						Attributes:  httpConfigAttributes(true),
					},
					"multistep": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Multi-step rule configuration. Used when rule_type is 'multistep'. Executes an ordered chain of HTTP requests, extracting variables from earlier responses for use in later steps.",
						Attributes: map[string]schema.Attribute{
							"steps": schema.ListNestedAttribute{
								Required:    true,
								Description: "Ordered list of HTTP requests to execute (1-20 steps).",
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required:    true,
											Description: "Human-readable name for the step. Surfaced in run history and error messages.",
										},
										"request": schema.SingleNestedAttribute{
											Required:    true,
											Description: "HTTP request configuration for this step.",
											Attributes:  httpConfigAttributes(false),
										},
										"extract": schema.ListNestedAttribute{
											Optional:    true,
											Description: "Rules that pull values from this step's response into named variables for use in later steps.",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Required:    true,
														Description: "Variable name. Must be uppercase, start with a letter, and be at least 3 characters (e.g., 'ACCESS_TOKEN').",
													},
													"source": schema.StringAttribute{
														Required:    true,
														Description: "Where to read the value from. Possible values: 'body', 'header'.",
														Validators: []validator.String{
															validatorutils.NewChoiceValidator(validExtractSources),
														},
													},
													"parser": schema.StringAttribute{
														Required:    true,
														Description: "How to extract the value. Possible values: 'jsonpath' (source 'body'), 'regex' (source 'body' or 'header'), 'header_value' (source 'header').",
														Validators: []validator.String{
															validatorutils.NewChoiceValidator(validExtractParsers),
														},
													},
													"query": schema.StringAttribute{
														Required:    true,
														Description: "The JSONPath expression, regex (first capture group), or header name to read.",
													},
													"secret": schema.BoolAttribute{
														Optional:    true,
														Computed:    true,
														Description: "Whether to redact the extracted value in results and logs. Defaults to false.",
													},
												},
											},
										},
										"continue_on_failure": schema.BoolAttribute{
											Optional:    true,
											Computed:    true,
											Description: "If true, a failing step does not stop the chain. Defaults to false.",
										},
										"exit_on_success": schema.BoolAttribute{
											Optional:    true,
											Computed:    true,
											Description: "If true, a successful step ends the chain early and marks the monitor as passed. Defaults to false.",
										},
									},
								},
							},
						},
					},
				},
			},
			"interval": schema.StringAttribute{
				Required:    true,
				CustomType:  validatorutils.NewDurationType(),
				Description: "Interval between checks (e.g., '30s', '1m').",
				Validators: []validator.String{
					validatorutils.NewDurationValidator(),
				},
			},
			"timeout": schema.StringAttribute{
				Required:    true,
				CustomType:  validatorutils.NewDurationType(),
				Description: "Timeout for each check (e.g., '5s', '10s').",
				Validators: []validator.String{
					validatorutils.NewDurationValidator(),
				},
			},
		},
	}
}
