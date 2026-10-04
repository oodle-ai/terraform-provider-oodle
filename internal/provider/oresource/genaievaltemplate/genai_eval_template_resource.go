package genaievaltemplate

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
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
	_ resource.Resource                   = &genaiEvalTemplateResource{}
	_ resource.ResourceWithConfigure      = &genaiEvalTemplateResource{}
	_ resource.ResourceWithImportState    = &genaiEvalTemplateResource{}
	_ resource.ResourceWithValidateConfig = &genaiEvalTemplateResource{}
)

var validTemplateTypes = map[string]struct{}{
	"llm":             {},
	"code":            {},
	"output_comparer": {},
}

// managedTemplatePrefix starts the id of each Oodle-managed template.
// The API does not accept a change of a managed template.
const managedTemplatePrefix = "oodle-managed-"

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

// ImportState refuses an Oodle-managed template. The API does not accept
// a change or a delete of one, thus each later apply or destroy would
// fail. An evaluator uses a managed template by its id instead.
func (r *genaiEvalTemplateResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	if strings.HasPrefix(req.ID, managedTemplatePrefix) {
		resp.Diagnostics.AddError(
			"Cannot import an Oodle-managed eval template",
			fmt.Sprintf(
				"%q is an Oodle-managed template, which cannot be "+
					"changed. Set it as the eval_template_id of an "+
					"oodle_genai_evaluator instead of importing it.",
				req.ID,
			),
		)

		return
	}

	r.APIBaseResource.ImportState(ctx, req, resp)
}

// ValidateConfig refuses params and library_pins on a template that is
// not 'code' at plan time. The API refuses them too, but only when the
// apply has started.
func (r *genaiEvalTemplateResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config genaiEvalTemplateResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Type.IsNull() || config.Type.IsUnknown() ||
		config.Type.ValueString() == codeTemplateType {
		return
	}

	if !config.Params.IsNull() && config.Params.ValueString() != "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("params"),
			"params is only for 'code' templates",
			"Remove params, or set type to 'code'.",
		)
	}

	if !config.LibraryPins.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("library_pins"),
			"library_pins is only for 'code' templates",
			"Remove library_pins, or set type to 'code'.",
		)
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
				CustomType: resourceutils.JSONType{},
				Optional:   true,
				Description: "JSON object describing the score the " +
					"template produces, for example " +
					"{\"score\": \"0 to 1\", \"reasoning\": \"why\"}.",
			},
			"model_params": schema.StringAttribute{
				CustomType: resourceutils.JSONType{},
				Optional:   true,
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
			"params": schema.StringAttribute{
				CustomType: resourceutils.JSONType{},
				Optional:   true,
				Description: "JSON array of the settings a 'code' " +
					"template declares. Each evaluator on the template " +
					"sets its own values for them in its params, and " +
					"the code reads them from ctx.params. Each setting " +
					"is an object with `name` (a lower-case Python " +
					"identifier), `type` (one of string, text, number, " +
					"integer, boolean, string_list, enum, json), and " +
					"optionally `label`, " +
					"`description`, `default`, `required` and, for an " +
					"enum, `options`. Only for 'code' templates.",
			},
			"library_pins": schema.MapAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
				Description: "Map of shared code library name to the " +
					"version of it this template runs, for example " +
					"{ text_utils = 2 }. A library that is not pinned " +
					"runs at its latest version. Use the name and " +
					"version attributes of an oodle_genai_code_library " +
					"so that Terraform creates the library first. " +
					"Only for 'code' templates.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Server-assigned version of the template.",
			},
		},
	}
}
