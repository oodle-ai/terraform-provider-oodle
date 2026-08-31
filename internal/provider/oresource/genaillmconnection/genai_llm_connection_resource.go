package genaillmconnection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &genaiLLMConnectionResource{}
	_ resource.ResourceWithConfigure   = &genaiLLMConnectionResource{}
	_ resource.ResourceWithImportState = &genaiLLMConnectionResource{}
)

// genaiLLMConnectionResource is the resource implementation.
type genaiLLMConnectionResource struct {
	oresource.APIBaseResource[
		*clientmodels.GenAILLMConnection,
		*genaiLLMConnectionResourceModel,
	]
}

func NewGenAILLMConnectionResource() resource.Resource {
	modelCreator := func() *clientmodels.GenAILLMConnection {
		return &clientmodels.GenAILLMConnection{}
	}

	return &genaiLLMConnectionResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.GenAILLMConnection,
			*genaiLLMConnectionResourceModel,
		](
			func() *genaiLLMConnectionResourceModel {
				return &genaiLLMConnectionResourceModel{}
			},
			modelCreator,
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.GenAILLMConnection] {
				return oodlehttp.NewGenAILLMConnectionClient(oodleHttpClient)
			},
		),
	}
}

func (r *genaiLLMConnectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_genai_llm_connection"
}

// Schema defines the schema for the resource.
func (r *genaiLLMConnectionResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "An LLM provider connection used by GenAI evaluators " +
			"and experiments to call a model.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the LLM connection.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the connection.",
			},
			// Named llm_provider because "provider" is a reserved
			// meta-argument in a Terraform resource block.
			"llm_provider": schema.StringAttribute{
				Required: true,
				Description: "LLM provider identifier, for example " +
					"'openai', 'anthropic' or 'bedrock'.",
			},
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				Description: "API key for the provider. The API stores this " +
					"encrypted and never returns it, so Terraform cannot " +
					"detect a key changed outside of Terraform.",
			},
			"base_url": schema.StringAttribute{
				Optional: true,
				Description: "Base URL to call instead of the provider's " +
					"default endpoint.",
			},
			"default_model": schema.StringAttribute{
				Optional:    true,
				Description: "Model used when a caller does not name one.",
			},
			"custom_models": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Additional model names to expose for this " +
					"connection.",
			},
			"enable_default_models": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "Whether the provider's built-in model list is " +
					"offered alongside custom_models.",
			},
			"is_default": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "Whether this is the connection used when none " +
					"is specified.",
			},
			"custom_headers": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Optional:   true,
				Sensitive:  true,
				Description: "JSON object of extra HTTP headers to send to " +
					"the provider. Stored encrypted and never returned.",
			},
			"default_params": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Optional:   true,
				Description: "JSON object of default model parameters, for " +
					"example {\"temperature\": 0}.",
			},
		},
	}
}
