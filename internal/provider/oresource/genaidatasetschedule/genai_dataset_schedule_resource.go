package genaidatasetschedule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	_ resource.Resource                   = &genaiDatasetScheduleResource{}
	_ resource.ResourceWithConfigure      = &genaiDatasetScheduleResource{}
	_ resource.ResourceWithImportState    = &genaiDatasetScheduleResource{}
	_ resource.ResourceWithValidateConfig = &genaiDatasetScheduleResource{}
)

// Schedule modes, matching the server's spelling.
const (
	ModeCalendar = "calendar"
	ModeInterval = "interval"
)

// defaultTimezone is what the server stores when a calendar schedule
// names none.
const defaultTimezone = "UTC"

var validModes = map[string]struct{}{
	ModeCalendar: {},
	ModeInterval: {},
}

var validIntervalUnits = map[string]struct{}{
	"minutes": {},
	"hours":   {},
	"days":    {},
}

// genaiDatasetScheduleResource is the resource implementation.
type genaiDatasetScheduleResource struct {
	oresource.APIBaseResource[
		*clientmodels.GenAIDatasetSchedule,
		*genaiDatasetScheduleResourceModel,
	]
}

func NewGenAIDatasetScheduleResource() resource.Resource {
	modelCreator := func() *clientmodels.GenAIDatasetSchedule {
		return &clientmodels.GenAIDatasetSchedule{}
	}

	return &genaiDatasetScheduleResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.GenAIDatasetSchedule,
			*genaiDatasetScheduleResourceModel,
		](
			func() *genaiDatasetScheduleResourceModel {
				return &genaiDatasetScheduleResourceModel{}
			},
			modelCreator,
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.GenAIDatasetSchedule] {
				return oodlehttp.NewGenAIDatasetScheduleClient(oodleHttpClient)
			},
		),
	}
}

func (r *genaiDatasetScheduleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_genai_dataset_schedule"
}

// Schema defines the schema for the resource.
func (r *genaiDatasetScheduleResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Runs a dataset's experiment on a schedule, so a " +
			"regression is found when a prompt or model change lands " +
			"rather than when the next person looks.\n\n" +
			"A dataset carries at most one schedule, so this is a " +
			"singleton keyed by dataset_name — declaring two for the " +
			"same dataset makes them fight over one row. The resource " +
			"is imported by dataset name.\n\n" +
			"A firing starts shortly after it is due rather than " +
			"exactly on the minute, and a schedule that falls behind " +
			"runs once instead of replaying every firing it missed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "Name of the dataset the schedule belongs " +
					"to. The schedule is addressed by dataset name, so " +
					"that is the id rather than schedule_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset_name": schema.StringAttribute{
				Required: true,
				Description: "Name of the dataset whose experiment this " +
					"schedule runs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required: true,
				Description: "Whether the schedule fires. False keeps " +
					"the definition but stops it running.",
			},
			"mode": schema.StringAttribute{
				Optional: true,
				Description: "Schedule shape: 'calendar' (the default) " +
					"reads timezone, times, weekdays and days_of_month; " +
					"'interval' reads interval_value and interval_unit.\n\n" +
					"The two answer different questions. Use calendar " +
					"for a run that has to land at a time of day someone " +
					"cares about, and interval for one whose cadence " +
					"matters but whose wall-clock time does not.",
				Validators: []validator.String{
					validatorutils.NewChoiceValidator(validModes),
				},
			},
			"interval_value": schema.Int64Attribute{
				Optional: true,
				Description: "How many interval_units between runs. The " +
					"period must be at least 5 minutes, which is what " +
					"the worker's poll cycle can honour, and at most " +
					"365 days. Interval mode only.",
			},
			"interval_unit": schema.StringAttribute{
				Optional: true,
				Description: "Unit of interval_value: 'minutes', " +
					"'hours' or 'days'. Interval mode only.",
				Validators: []validator.String{
					validatorutils.NewChoiceValidator(validIntervalUnits),
				},
			},
			"timezone": schema.StringAttribute{
				Optional: true,
				Description: "IANA timezone the times are read in, " +
					"defaulting to UTC. Holding the times to a named " +
					"zone is what keeps a schedule on the wall clock " +
					"across a daylight saving change rather than " +
					"drifting by an hour twice a year. Calendar mode " +
					"only: a duration is the same length everywhere.",
			},
			"times": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Times of day the experiment starts, as " +
					"\"HH:MM\" in timezone. At least one is required in " +
					"calendar mode.",
			},
			"weekdays": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Lowercase weekday names to run on, e.g. " +
					"[\"monday\", \"friday\"]. Empty means every day. " +
					"Calendar mode only.",
			},
			"days_of_month": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Days of the month to run on, \"1\" to " +
					"\"31\". Empty means every day, and with weekdays " +
					"also set a day has to satisfy both. Calendar mode " +
					"only.",
			},
			"experiment_config": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
				Description: "JSON object describing what each firing " +
					"runs — the same llm-experiment job config the API " +
					"takes, with datasetId, llmConnectionId, a prompt " +
					"(promptName or promptTemplate), and optional " +
					"evaluatorIds, outputComparerIds and " +
					"evalConnectionId.\n\n" +
					"runName is ignored: every firing is numbered on " +
					"its own. An evaluator id has to sit in the list " +
					"matching its template type — an output comparer " +
					"under outputComparerIds, anything else under " +
					"evaluatorIds — or the run is rejected.",
			},
			"schedule_id": schema.StringAttribute{
				Computed:    true,
				Description: "Server-assigned uuid of the schedule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset_id": schema.StringAttribute{
				Computed: true,
				Description: "Server-assigned uuid of the dataset the " +
					"schedule belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"next_run_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the schedule fires next.",
			},
			"last_run_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the schedule last fired.",
			},
			"last_error": schema.StringAttribute{
				Computed: true,
				Description: "Why the most recent launch did not start, " +
					"and empty once one succeeds. A scheduled run that " +
					"cannot be queued produces no job and no dataset " +
					"run, so this is the only place the failure shows.",
			},
		},
	}
}

// ValidateConfig reports the field-group mistakes at plan time.
//
// Which fields are required depends on mode, which a schema cannot
// express, so the server checks it. Repeating the check here turns a
// half-applied plan into a plan that never starts, and the wrong-mode
// fields are worth naming: the server ignores them silently, so a
// schedule that fires at a cadence nobody asked for looks correct in
// the config that produced it.
func (r *genaiDatasetScheduleResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var model genaiDatasetScheduleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// An unknown mode decides nothing: which fields are required
	// depends on it, so judging one branch would report the other
	// branch's fields as missing. `for_each` is not expanded at
	// validate time, so this is an ordinary config rather than a
	// broken one.
	if model.Mode.IsUnknown() {
		return
	}

	// Empty mode means calendar, the same default the server applies.
	mode := ModeCalendar
	if !model.Mode.IsNull() && model.Mode.ValueString() != "" {
		mode = model.Mode.ValueString()
	}

	hasInterval := !model.IntervalValue.IsNull() ||
		!model.IntervalUnit.IsNull()
	hasCalendar := hasElements(model.Times) ||
		hasElements(model.Weekdays) ||
		hasElements(model.DaysOfMonth) ||
		!model.Timezone.IsNull()

	if mode == ModeInterval {
		if model.IntervalValue.IsNull() || model.IntervalUnit.IsNull() {
			resp.Diagnostics.AddError(
				"Incomplete interval schedule",
				"An interval schedule needs both interval_value and "+
					"interval_unit, for example 6 and \"hours\".",
			)
		}
		if hasCalendar {
			resp.Diagnostics.AddError(
				"Calendar fields on an interval schedule",
				"timezone, times, weekdays and days_of_month apply "+
					"only to a calendar schedule. The server ignores "+
					"them here, so leave them out or set "+
					"mode = \"calendar\".",
			)
		}

		return
	}

	if !hasElements(model.Times) {
		resp.Diagnostics.AddAttributeError(
			path.Root("times"),
			"Calendar schedule without times",
			"A calendar schedule needs at least one time, for example "+
				"[\"09:00\"]. Set mode = \"interval\" and "+
				"interval_value/interval_unit for a schedule that runs "+
				"on a period instead.",
		)
	}
	if hasInterval {
		resp.Diagnostics.AddError(
			"Interval fields on a calendar schedule",
			"interval_value and interval_unit apply only to an "+
				"interval schedule. The server ignores them here, so "+
				"leave them out or set mode = \"interval\".",
		)
	}
}

// hasElements reports whether a list attribute holds anything. An
// unknown list is one built from another resource's output, which
// cannot be judged until apply, so it counts as set.
func hasElements(list types.List) bool {
	if list.IsUnknown() {
		return true
	}
	if list.IsNull() {
		return false
	}

	return len(list.Elements()) > 0
}
