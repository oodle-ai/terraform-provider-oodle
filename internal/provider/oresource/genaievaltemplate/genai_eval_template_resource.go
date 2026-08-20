package genaievaltemplate

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
	"terraform-provider-oodle/internal/validatorutils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &genaiEvalTemplateResource{}
	_ resource.ResourceWithConfigure   = &genaiEvalTemplateResource{}
	_ resource.ResourceWithImportState = &genaiEvalTemplateResource{}
)

var validTemplateTypes = map[string]struct{}{
	"llm":             {},
	"code":            {},
	"output_comparer": {},
}

// genaiEvalTemplateResource is the resource implementation.
type genaiEvalTemplateResource struct {
	oresource.APIBaseResource[
		*clientmodels.GenAIEvalTemplate,
		*genaiEvalTemplateResourceModel,
	]
}

func NewGenAIEvalTemplateResource() resource.Resource {
	modelCreator := func() *clientmodels.GenAIEvalTemplate {
		return &clientmodels.GenAIEvalTemplate{}
	}

	return &genaiEvalTemplateResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.GenAIEvalTemplate,
			*genaiEvalTemplateResourceModel,
		](
			func() *genaiEvalTemplateResourceModel {
				return &genaiEvalTemplateResourceModel{}
			},
			modelCreator,
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.GenAIEvalTemplate] {
				return oodlehttp.NewGenAIEvalTemplateClient(oodleHttpClient)
			},
		),
	}
}

func (r *genaiEvalTemplateResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_genai_eval_template"
}

// Schema defines the schema for the resource.
func (r *genaiEvalTemplateResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "A reusable GenAI scoring definition — an " +
			"LLM-as-judge prompt, a code scorer, or an output comparer " +
			"that scores against a dataset item's expected output. This " +
			"is what the Oodle UI calls a Library template. Attach it " +
			"to traffic with an oodle_genai_evaluator resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the eval template.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the eval template.",
			},
			"type": schema.StringAttribute{
				Required: true,
				Description: "Kind of scorer: 'llm' for an LLM-as-judge " +
					"prompt, 'code' for a Python scorer, or " +
					"'output_comparer' for a judge that scores the " +
					"output against a dataset item's expected output. " +
					"Cannot be changed after creation.\n\n" +
					"An output comparer's prompt uses {{output}} and " +
					"{{expected_output}}. Ground truth only exists " +
					"inside an experiment, so a comparer never runs " +
					"against live traffic: an evaluator built on one " +
					"produces scores only through an experiment run, " +
					"and an item with no expected output is skipped " +
					"rather than scored zero.",
				Validators: []validator.String{
					validatorutils.NewChoiceValidator(validTemplateTypes),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt": schema.StringAttribute{
				Optional: true,
				Description: "Judge prompt for 'llm' templates, using " +
					"{{var}} placeholders drawn from vars.",
			},
			"vars": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Variable names the prompt or source code " +
					"expects. An evaluator maps span fields onto these.",
			},
			"output_schema": schema.StringAttribute{
				Optional: true,
				Description: "JSON object describing the score the " +
					"template produces, for example " +
					"{\"score\": \"0 to 1\", \"reasoning\": \"why\"}.",
			},
			"model_params": schema.StringAttribute{
				Optional: true,
				Description: "JSON object of model parameters used when " +
					"running the judge, for example {\"temperature\": 0}.",
			},
			"source_code": schema.StringAttribute{
				Optional: true,
				Description: "Python source for 'code' templates. Requires " +
					"an enterprise plan and the code evaluator feature.",
			},
			"source_code_language": schema.StringAttribute{
				Optional:    true,
				Description: "Language of source_code. Only 'python' today.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Server-assigned version of the template.",
			},
		},
	}
}
