package mutingrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	_ resource.Resource                   = &mutingRuleResource{}
	_ resource.ResourceWithConfigure      = &mutingRuleResource{}
	_ resource.ResourceWithImportState    = &mutingRuleResource{}
	_ resource.ResourceWithValidateConfig = &mutingRuleResource{}
)

// monitorIDLabel is the label a muting rule must pin to a monitor.
const monitorIDLabel = "_oodle_monitor_id"

var validMatchTypes = map[string]struct{}{
	"=":  {},
	"!=": {},
	"=~": {},
	"!~": {},
}

// mutingRuleResource is the resource implementation.
type mutingRuleResource struct {
	oresource.APIBaseResource[
		*clientmodels.MutingRule,
		*mutingRuleResourceModel,
	]
}

func NewMutingRuleResource() resource.Resource {
	modelCreator := func() *clientmodels.MutingRule {
		return &clientmodels.MutingRule{}
	}

	return &mutingRuleResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.MutingRule,
			*mutingRuleResourceModel,
		](
			func() *mutingRuleResourceModel {
				return &mutingRuleResourceModel{}
			},
			modelCreator,
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.MutingRule] {
				return oodlehttp.NewMutingRuleClient(oodleHttpClient)
			},
		),
	}
}

func (r *mutingRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_muting_rule"
}

// Schema defines the schema for the resource.
func (r *mutingRuleResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Suppresses notifications for the alerts its matchers " +
			"select.\n\n" +
			"A rule is either one-off, bounded by `starts_at` and " +
			"`ends_at`, or recurring, driven by `schedule_ids`. The two " +
			"are mutually exclusive.\n\n" +
			"Muting is scoped to a monitor, so every rule must include an " +
			"equality matcher on `" + monitorIDLabel + "`.\n\n" +
			"Every attribute forces replacement. A one-off rule is " +
			"backed by an Alertmanager silence, and editing one that is " +
			"already active makes Alertmanager mint a new silence id " +
			"while the API still reports the old one — so an in-place " +
			"update would leave Terraform holding a stale id and orphan " +
			"the rule it just wrote. Replacing keeps state and server " +
			"in step.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the muting rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional: true,
				Description: "Human-readable name for the muting rule. " +
					"Recurring rules only: a one-off rule is stored as an " +
					"Alertmanager silence, which has no name field, so the " +
					"API would silently drop it. Use comment instead.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Optional: true,
				Description: "Note recorded alongside the rule, typically " +
					"why the alerts are muted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"matchers": schema.ListNestedAttribute{
				Required: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Description: "Label matchers selecting the alerts to mute. " +
					"An alert is muted when it matches all of them. One " +
					"must be an equality matcher on `" + monitorIDLabel +
					"`, naming the monitor being muted.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
							Description: "The type of match to perform. " +
								"Valid values are: '=' (equals), " +
								"'!=' (not equals), '=~' (regex match), " +
								"'!~' (regex not match).",
							Validators: []validator.String{
								validatorutils.NewChoiceValidator(
									validMatchTypes,
								),
							},
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the label to match against.",
						},
						"value": schema.StringAttribute{
							Required: true,
							Description: "The value to match against. For " +
								"regex matches, this must be a valid " +
								"regular expression.",
						},
					},
				},
			},
			"starts_at": schema.StringAttribute{
				Optional: true,
				// Computed because the API fills this in with the
				// creation time when it is left out.
				Computed: true,
				Description: "RFC 3339 timestamp when muting begins. " +
					"Defaults to the time the rule is created. One-off " +
					"rules only.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ends_at": schema.StringAttribute{
				Optional: true,
				Description: "RFC 3339 timestamp when muting ends. Muting " +
					"is effective forever if unset, and the API rejects a " +
					"time in the past. One-off rules only.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schedule_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "IDs of the `oodle_muting_schedule`s during " +
					"which the alert is muted. Setting this makes the " +
					"rule recurring, which today supports exactly one " +
					"matcher — the `" + monitorIDLabel + "` one — and " +
					"cannot be combined with starts_at or ends_at.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the muting rule.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the muting rule was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the muting rule was last updated.",
			},
		},
	}
}

// ValidateConfig reports the two constraints the API enforces, at
// plan time rather than half way through an apply.
func (r *mutingRuleResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var model mutingRuleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasSchedules := !model.ScheduleIDs.IsNull() &&
		!model.ScheduleIDs.IsUnknown() &&
		len(model.ScheduleIDs.Elements()) > 0
	hasTimeRange := isSet(model.StartsAt) || isSet(model.EndsAt)

	if hasSchedules && hasTimeRange {
		resp.Diagnostics.AddError(
			"Conflicting muting rule schedule",
			"A muting rule is either recurring, with schedule_ids, or "+
				"one-off, with starts_at and ends_at. Set one or the "+
				"other, not both.",
		)
	}

	// A one-off rule is stored as an Alertmanager silence, which has
	// no name field. Accepting a name would mean writing it to state
	// and never to the server, so it is refused outright rather than
	// disappearing on the next read.
	if !hasSchedules && isSet(model.Name) {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Name is not stored for one-off muting rules",
			"A one-off muting rule is stored as an Alertmanager "+
				"silence, which has no name field, so the API drops "+
				"the name. Use comment for a one-off rule, or set "+
				"schedule_ids to make the rule recurring.",
		)
	}

	// Only a fully known matcher list can be judged; one built from
	// another resource's id is unknown until apply.
	if model.Matchers.IsNull() || model.Matchers.IsUnknown() {
		return
	}

	matchers, err := matchersToClient(ctx, model.Matchers)
	if err != nil {
		return
	}

	for _, matcher := range matchers {
		if matcher.Name == monitorIDLabel &&
			matcher.Type.String() == "=" {
			return
		}
	}

	resp.Diagnostics.AddError(
		"Muting rule is not scoped to a monitor",
		"matchers must include an equality matcher on \""+
			monitorIDLabel+"\", naming the monitor being muted. "+
			"For example:\n\n"+
			"  matchers = [{\n"+
			"    type  = \"=\"\n"+
			"    name  = \""+monitorIDLabel+"\"\n"+
			"    value = oodle_monitor.example.id\n"+
			"  }]",
	)
}

func isSet(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown() && value.ValueString() != ""
}
