package genaievaluator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/resourceutils"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
	"terraform-provider-oodle/internal/validatorutils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &genaiEvaluatorResource{}
	_ resource.ResourceWithConfigure   = &genaiEvaluatorResource{}
	_ resource.ResourceWithImportState = &genaiEvaluatorResource{}
)

// genaiEvaluatorResource is the resource implementation.
type genaiEvaluatorResource struct {
	oresource.APIBaseResource[
		*clientmodels.GenAIEvaluationRule,
		*genaiEvaluatorResourceModel,
	]
}

func NewGenAIEvaluatorResource() resource.Resource {
	modelCreator := func() *clientmodels.GenAIEvaluationRule {
		return &clientmodels.GenAIEvaluationRule{}
	}

	return &genaiEvaluatorResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.GenAIEvaluationRule,
			*genaiEvaluatorResourceModel,
		](
			func() *genaiEvaluatorResourceModel {
				return &genaiEvaluatorResourceModel{}
			},
			modelCreator,
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.GenAIEvaluationRule] {
				return oodlehttp.NewGenAIEvaluationRuleClient(oodleHttpClient)
			},
		),
	}
}

func (r *genaiEvaluatorResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_genai_evaluator"
}

// Schema defines the schema for the resource.
func (r *genaiEvaluatorResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Runs a GenAI eval template against live traffic: " +
			"which spans it scores, how often, and how span fields map " +
			"onto the template's variables.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the evaluator.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the evaluator.",
			},
			"eval_template_id": schema.StringAttribute{
				Required: true,
				Description: "ID of the eval template to run. Use the id of " +
					"an oodle_genai_eval_template, or the id of an " +
					"Oodle-managed template. Cannot be changed after " +
					"creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the evaluator scores new traffic.",
			},
			"target_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "What the evaluator runs against, for example " +
					"'trace'.",
			},
			"sampling_rate": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Description: "Fraction of matching spans to score, from 0 " +
					"to 1.",
			},
			"max_invocations_per_hour": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Description: "Ceiling on how many times the evaluator may " +
					"run per hour.",
			},
			"filters": schema.StringAttribute{
				CustomType: resourceutils.JSONType{},
				Optional:   true,
				Description: "JSON array of filters selecting which spans " +
					"are scored. An empty or unset value scores everything " +
					"that matches target_type.\n\n" +
					"Each filter is an object with `name`, `type` and " +
					"`value`, or `multi_value` for types 4 and 5. " +
					"`name` is `span::<attribute>` for a span " +
					"attribute, `resource::<attribute>` for a resource " +
					"attribute, or a plain column name such as " +
					"`service_name`. A dotted name without one of the two " +
					"prefixes is refused, because it does not say which " +
					"kind of attribute it is. `type` is the matcher in " +
					"the form the API stores: `0` (equals), `1` (not " +
					"equals), `2` (regex match), `3` (regex does not " +
					"match), `4` (one of), `5` (not one of), or " +
					"`\"GT\"`, `\"GTE\"`, `\"LT\"`, `\"LTE\"` for numeric " +
					"comparisons. Types 4 and 5 take their " +
					"values in a `multi_value` list instead of `value`. " +
					"Operator words such as " +
					"`\"eq\"` are refused at plan time, because the API " +
					"rewrites them and the stored value would never " +
					"match the configuration.\n\n" +
					"An evaluator that reads the scores of another one " +
					"must have the same filters as it; the API refuses " +
					"the pair otherwise.",
				Validators: []validator.String{
					validatorutils.NewGenAIFiltersValidator(),
				},
			},
			"variable_mapping": schema.StringAttribute{
				CustomType: resourceutils.JSONType{},
				Optional:   true,
				Description: "JSON describing how span fields populate the " +
					"eval template's vars.",
			},
			"params": schema.StringAttribute{
				CustomType: resourceutils.JSONType{},
				Optional:   true,
				Description: "JSON object of values for the settings " +
					"that a 'code' template declares in its params, by " +
					"setting name, for example " +
					"{\"threshold\": 0.8}. A setting left out runs with " +
					"the template's default. Only for evaluators on " +
					"'code' templates.",
			},
			"score_input_rule_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "IDs of the evaluators whose scores this " +
					"one reads. The API derives them from the template " +
					"on each write.",
			},
			"llm_connection_id": schema.StringAttribute{
				Optional: true,
				Description: "ID of the oodle_genai_llm_connection the " +
					"judge calls. Required for 'llm' templates.",
			},
			"model_params": schema.StringAttribute{
				CustomType: resourceutils.JSONType{},
				Optional:   true,
				Description: "JSON object of model parameters overriding " +
					"the template's own.",
			},
			"depends_on_rule_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "IDs of evaluators that must score a span " +
					"before this one runs.",
			},
			"dataset_id": schema.StringAttribute{
				Optional: true,
				Description: "Restricts the evaluator to runs over one " +
					"dataset. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
