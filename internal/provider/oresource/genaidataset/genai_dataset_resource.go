package genaidataset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &genaiDatasetResource{}
	_ resource.ResourceWithConfigure   = &genaiDatasetResource{}
	_ resource.ResourceWithImportState = &genaiDatasetResource{}
)

// genaiDatasetResource is the resource implementation.
type genaiDatasetResource struct {
	oresource.APIBaseResource[
		*clientmodels.GenAIDataset,
		*genaiDatasetResourceModel,
	]
}

func NewGenAIDatasetResource() resource.Resource {
	modelCreator := func() *clientmodels.GenAIDataset {
		return &clientmodels.GenAIDataset{}
	}

	return &genaiDatasetResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.GenAIDataset,
			*genaiDatasetResourceModel,
		](
			func() *genaiDatasetResourceModel {
				return &genaiDatasetResourceModel{}
			},
			modelCreator,
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.GenAIDataset] {
				return oodlehttp.NewGenAIDatasetClient(oodleHttpClient)
			},
		),
	}
}

func (r *genaiDatasetResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_genai_dataset"
}

// Schema defines the schema for the resource.
func (r *genaiDatasetResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "A named collection of GenAI evaluation inputs. Add " +
			"rows with oodle_genai_dataset_item.\n\n" +
			"Datasets have no update endpoint, so every attribute forces " +
			"replacement — and replacing a dataset deletes its items and " +
			"run history. The resource is imported by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "Name of the dataset. Datasets are addressed " +
					"by name rather than by uuid.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the dataset.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "What the dataset covers.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metadata": schema.StringAttribute{
				Optional:    true,
				Description: "JSON object of arbitrary metadata.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dataset_id": schema.StringAttribute{
				Computed: true,
				Description: "Server-assigned uuid of the dataset. Use this " +
					"where an evaluator or experiment asks for a dataset id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
