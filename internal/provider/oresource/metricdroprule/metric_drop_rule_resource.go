package metricdroprule

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
	_ resource.Resource                = &metricDropRuleResource{}
	_ resource.ResourceWithConfigure   = &metricDropRuleResource{}
	_ resource.ResourceWithImportState = &metricDropRuleResource{}
)

const dropRulesResourcePath = "drop-rules"

var validMatchTypes = map[string]struct{}{
	"=":  {},
	"!=": {},
	"=~": {},
	"!~": {},
}

// metricDropRuleResource is the resource implementation.
type metricDropRuleResource struct {
	oresource.BaseResource[*clientmodels.MetricDropRule, *metricDropRuleResourceModel]
}

func NewMetricDropRuleResource() resource.Resource {
	modelCreator := func() *clientmodels.MetricDropRule {
		return &clientmodels.MetricDropRule{}
	}
	return &metricDropRuleResource{
		BaseResource: oresource.NewBaseResource[*clientmodels.MetricDropRule, *metricDropRuleResourceModel](
			func() *metricDropRuleResourceModel {
				return &metricDropRuleResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[*clientmodels.MetricDropRule] {
				return oodlehttp.NewModelClient[*clientmodels.MetricDropRule](
					oodleHttpClient,
					dropRulesResourcePath,
					modelCreator,
				)
			},
		),
	}
}

func matcherSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "Label name to match against.",
		},
		"type": schema.StringAttribute{
			Required:    true,
			Description: "Match type. Possible values are: '=' (exact), '!=' (not equal), '=~' (regex), '!~' (negative regex).",
			Validators: []validator.String{
				validatorutils.NewChoiceValidator(validMatchTypes),
			},
		},
		"value": schema.StringAttribute{
			Required:    true,
			Description: "Value or pattern to match against.",
		},
	}
}

// Metadata returns the resource type name.
func (r *metricDropRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metric_drop_rule"
}

// Schema defines the schema for the resource.
func (r *metricDropRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a metric drop rule. Drop rules prevent specific metric time-series from being ingested into Oodle.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the metric drop rule.",
			},
			"rule_name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name for the drop rule.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Type of the drop rule.",
			},
			"metric_name": schema.SingleNestedAttribute{
				Required:    true,
				Attributes:  matcherSchemaAttributes(),
				Description: "The __name__ label matcher that selects which metrics to drop.",
			},
			"filters": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: matcherSchemaAttributes(),
				},
				Description: "Optional additional label matchers that further restrict which series are dropped.",
			},
		},
	}
}
