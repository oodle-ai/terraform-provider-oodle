package recordingrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
)

var (
	_ resource.Resource                = &recordingRuleResource{}
	_ resource.ResourceWithConfigure   = &recordingRuleResource{}
	_ resource.ResourceWithImportState = &recordingRuleResource{}
)

const recordingRulesResourcePath = "recording-rules"

type recordingRuleResource struct {
	oresource.BaseResource[*clientmodels.RecordingRuleEnvelope, *recordingRuleResourceModel]
}

func NewRecordingRuleResource() resource.Resource {
	modelCreator := func() *clientmodels.RecordingRuleEnvelope {
		return &clientmodels.RecordingRuleEnvelope{}
	}
	return &recordingRuleResource{
		BaseResource: oresource.NewBaseResource[*clientmodels.RecordingRuleEnvelope, *recordingRuleResourceModel](
			func() *recordingRuleResourceModel {
				return &recordingRuleResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[*clientmodels.RecordingRuleEnvelope] {
				return oodlehttp.NewModelClient[*clientmodels.RecordingRuleEnvelope](
					oodleHttpClient,
					recordingRulesResourcePath,
					modelCreator,
				)
			},
		),
	}
}

// Metadata returns the resource type name.
func (r *recordingRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_recording_rule"
}

// Schema defines the schema for the resource.
func (r *recordingRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a recording rule. A recording rule evaluates a PromQL expression on an " +
			"interval and stores the result as a new metric. Use recording rules to make an expensive " +
			"query cheap, or to record SLO error ratios that burn-rate monitors read. Import a rule " +
			"by its slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the recording rule. It is the same as the slug.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Unique ID of the rule in the instance. If you do not set it, Oodle makes " +
					"one from the name. A change of the slug replaces the rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the rule.",
			},
			"metric_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the metric that the rule writes.",
			},
			"prometheus_expr": schema.StringAttribute{
				Required: true,
				Description: "PromQL expression that the rule evaluates. It must return an instant " +
					"vector or a scalar.",
			},
			"interval_secs": schema.Int64Attribute{
				Optional: true,
				Description: "How often the rule runs, in seconds, from 60 to 3600. The default is 60. " +
					"All rules in one execution group run at the shortest interval of the group.",
			},
			"execution_group": schema.StringAttribute{
				Optional: true,
				Description: "Name of a set of rules that run together, one after another, in the " +
					"order of their creation. Put a rule that reads the metric of another rule in " +
					"the same group. If you do not set it, the rule runs in a group of its own.",
			},
			"labels": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Labels to set on every series that the rule writes.",
			},
		},
	}
}
