package genaidatasetitem

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"

	"terraform-provider-oodle/internal/resourceutils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &genaiDatasetItemResource{}
	_ resource.ResourceWithConfigure   = &genaiDatasetItemResource{}
	_ resource.ResourceWithImportState = &genaiDatasetItemResource{}
)

// genaiDatasetItemResource is the resource implementation.
type genaiDatasetItemResource struct {
	oresource.APIBaseResource[
		*clientmodels.GenAIDatasetItem,
		*genaiDatasetItemResourceModel,
	]
}

func NewGenAIDatasetItemResource() resource.Resource {
	modelCreator := func() *clientmodels.GenAIDatasetItem {
		return &clientmodels.GenAIDatasetItem{}
	}

	return &genaiDatasetItemResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.GenAIDatasetItem,
			*genaiDatasetItemResourceModel,
		](
			func() *genaiDatasetItemResourceModel {
				return &genaiDatasetItemResourceModel{}
			},
			modelCreator,
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.GenAIDatasetItem] {
				return oodlehttp.NewGenAIDatasetItemClient(oodleHttpClient)
			},
		),
	}
}

func (r *genaiDatasetItemResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_genai_dataset_item"
}

// Schema defines the schema for the resource.
func (r *genaiDatasetItemResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "One row of a GenAI dataset: an input, and optionally " +
			"the output it is expected to produce.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the dataset item.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset_name": schema.StringAttribute{
				Required: true,
				Description: "Name of the dataset this item belongs to. " +
					"Moving an item between datasets forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input": schema.StringAttribute{
				CustomType: resourceutils.JSONType{},
				Required:   true,
				Description: "JSON value given to the prompt under test. " +
					"Use jsonencode() to build it.",
			},
			"expected_output": schema.StringAttribute{
				CustomType: resourceutils.JSONType{},
				Optional:   true,
				Description: "JSON value the output is compared against by " +
					"evaluators that take a reference answer.",
			},
			"metadata": schema.StringAttribute{
				CustomType:  resourceutils.JSONType{},
				Optional:    true,
				Description: "JSON object of arbitrary metadata.",
			},
			"status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Whether the item is used by runs. 'ACTIVE' or " +
					"'ARCHIVED'.",
			},
			"source_trace_id": schema.StringAttribute{
				Optional: true,
				Description: "Trace this item was captured from, if any. " +
					"Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_observation_id": schema.StringAttribute{
				Optional: true,
				Description: "Observation within source_trace_id this item " +
					"was captured from. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dataset_id": schema.StringAttribute{
				Computed:    true,
				Description: "Server-assigned uuid of the parent dataset.",
			},
		},
	}
}
