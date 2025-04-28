package logmetrics

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
	"terraform-provider-oodle/internal/validatorutils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &logMetricsResource{}
	_ resource.ResourceWithConfigure   = &logMetricsResource{}
	_ resource.ResourceWithImportState = &logMetricsResource{}
)

const logMetricsResourceName = "logmetrics"

var validOperators = map[string]struct{}{
	"is":            {},
	"contains":      {},
	"matches regex": {},
	"exists":        {},
}

var validMetricTypes = map[string]struct{}{
	"log_count": {},
	"counter":   {},
	"gauge":     {},
	"histogram": {},
}

// logMetricsResource is the resource implementation.
type logMetricsResource struct {
	oresource.BaseResource[*clientmodels.LogMetrics, *logMetricsResourceModel]
}

func NewLogMetricsResource() resource.Resource {
	modelCreator := func() *clientmodels.LogMetrics {
		return &clientmodels.LogMetrics{}
	}
	return &logMetricsResource{
		BaseResource: oresource.NewBaseResource[*clientmodels.LogMetrics, *logMetricsResourceModel](
			func() *logMetricsResourceModel {
				return &logMetricsResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[*clientmodels.LogMetrics] {
				return oodlehttp.NewModelClient[*clientmodels.LogMetrics](
					oodleHttpClient,
					logMetricsResourceName,
					modelCreator,
				)
			},
		),
	}
}

// getFilterSchema returns the schema for a filter.
func getFilterSchema() map[string]schema.Attribute {
	matchSchema := map[string]schema.Attribute{
		"field": schema.StringAttribute{
			Required:    true,
			Description: "Name of the log field to match against.",
		},
		"json_path": schema.StringAttribute{
			Optional:    true,
			Description: "JSONPath to match against a value at a specific path in the JSON field.",
		},
		"operator": schema.StringAttribute{
			Required:    true,
			Description: "Operator to use for matching. Possible values are: 'is', 'contains', 'matches regex', 'exists'.",
			Validators: []validator.String{
				validatorutils.NewChoiceValidator(validOperators),
			},
		},
		"value": schema.StringAttribute{
			Optional:    true,
			Description: "Value to match against.",
		},
	}

	// Allow only match and not within all filters
	allNestedFilterSchema := map[string]schema.Attribute{
		"match": schema.SingleNestedAttribute{
			Optional:    true,
			Attributes:  matchSchema,
			Description: "Simple field matching filter.",
		},
		"not": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"match": schema.SingleNestedAttribute{
					Optional:    true,
					Attributes:  matchSchema,
					Description: "Simple field matching filter.",
				},
			},
			Description: "Filter that must not match.",
		},
	}

	// Allow match, not and all within any filters
	anyNestedFilterSchema := map[string]schema.Attribute{
		"match": schema.SingleNestedAttribute{
			Optional:    true,
			Attributes:  matchSchema,
			Description: "Simple field matching filter.",
		},
		"not": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"match": schema.SingleNestedAttribute{
					Optional:    true,
					Attributes:  matchSchema,
					Description: "Simple field matching filter.",
				},
			},
			Description: "Filter that must not match.",
		},
		"all": schema.ListNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: allNestedFilterSchema,
			},
			Description: "List of filters where all must match.",
		},
	}

	// Allow only match within not filter
	notNestedFilterSchema := map[string]schema.Attribute{
		"match": schema.SingleNestedAttribute{
			Optional:    true,
			Attributes:  matchSchema,
			Description: "Simple field matching filter.",
		},
	}

	// Define the top-level filter schema
	filterSchema := map[string]schema.Attribute{
		"match": schema.SingleNestedAttribute{
			Optional:    true,
			Attributes:  matchSchema,
			Description: "Simple field matching filter.",
		},
		"all": schema.ListNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: allNestedFilterSchema,
			},
			Description: "List of filters where all must match.",
		},
		"any": schema.ListNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: anyNestedFilterSchema,
			},
			Description: "List of filters where at least one must match.",
		},
		"not": schema.SingleNestedAttribute{
			Optional:    true,
			Attributes:  notNestedFilterSchema,
			Description: "Filter that must not match.",
		},
	}

	return filterSchema
}

// Metadata returns the resource type name.
func (r *logMetricsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logmetrics"
}

// Schema defines the schema for the resource.
func (r *logMetricsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the log metrics rule.",
				Validators: []validator.String{
					validatorutils.NewUUIDValidator(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the log metrics rule.",
			},
			"labels": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the label.",
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Description: "Static value of the label. Only one of value or value_extractor should be set.",
						},
						"value_extractor": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"field": schema.StringAttribute{
									Optional:    true,
									Description: "Name of the field in the log to extract the value from.",
								},
								"json_path": schema.StringAttribute{
									Optional:    true,
									Description: "JSONPath to extract a nested value from a JSON field.",
								},
								"regex": schema.StringAttribute{
									Optional:    true,
									Description: "Regex pattern to extract a value from the field.",
								},
							},
							Description: "Configuration for extracting label values from log fields.",
						},
					},
				},
				Description: "Labels to be added to all metrics created by this configuration.",
			},
			"filter": schema.SingleNestedAttribute{
				Optional:    true,
				Attributes:  getFilterSchema(),
				Description: "Filter to determine which logs to process.",
				Validators: []validator.Object{
					validatorutils.NewFilterValidator(),
				},
			},
			"metric_definitions": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the metric to be created. Must match Prometheus metric naming rules.",
							Validators: []validator.String{
								validatorutils.NewMetricNameValidator(),
							},
						},
						"type": schema.StringAttribute{
							Required: true,
							Description: "Type of metric to create. Possible values are:\n" +
								"  - `log_count` - Counts the number of logs that match the filter.\n" +
								"  - `counter` - Extracts and sums numeric values from fields.\n" +
								"  - `gauge` - Records the latest numeric value from fields.\n" +
								"  - `histogram` - Creates distribution buckets of numeric values from fields.",
							Validators: []validator.String{
								validatorutils.NewChoiceValidator(validMetricTypes),
							},
						},
						"field": schema.StringAttribute{
							Optional:    true,
							Description: "Name of the log field to extract from. Only used when type is not 'log_count'.",
						},
						"json_path": schema.StringAttribute{
							Optional:    true,
							Description: "JSONPath to extract a numeric value from a JSON field. Cannot be used together with regex.",
						},
						"regex": schema.StringAttribute{
							Optional:    true,
							Description: "Regex pattern to extract a numeric value from the field. Cannot be used together with json_path.",
						},
					},
				},
				Description: "Definitions of metrics to be created from the logs.",
			},
		},
	}
}
