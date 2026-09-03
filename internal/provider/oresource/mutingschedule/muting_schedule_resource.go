package mutingschedule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &mutingScheduleResource{}
	_ resource.ResourceWithConfigure      = &mutingScheduleResource{}
	_ resource.ResourceWithImportState    = &mutingScheduleResource{}
	_ resource.ResourceWithValidateConfig = &mutingScheduleResource{}
)

// mutingScheduleResource is the resource implementation.
type mutingScheduleResource struct {
	oresource.APIBaseResource[
		*clientmodels.MutingSchedule,
		*mutingScheduleResourceModel,
	]
}

func NewMutingScheduleResource() resource.Resource {
	return &mutingScheduleResource{
		APIBaseResource: oresource.NewAPIBaseResource[
			*clientmodels.MutingSchedule,
			*mutingScheduleResourceModel,
		](
			func() *mutingScheduleResourceModel {
				return &mutingScheduleResourceModel{}
			},
			func() *clientmodels.MutingSchedule {
				return &clientmodels.MutingSchedule{}
			},
			func(
				oodleHttpClient *oodlehttp.OodleApiClient,
			) oresource.ModelAPI[*clientmodels.MutingSchedule] {
				return oodlehttp.NewMutingScheduleClient(oodleHttpClient)
			},
		),
	}
}

func (r *mutingScheduleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_muting_schedule"
}

// Schema defines the schema for the resource.
func (r *mutingScheduleResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "A recurring window that muting rules mute on.\n\n" +
			"A schedule mutes nothing by itself. Reference it from an " +
			"`oodle_muting_rule`'s `schedule_ids` to mute that rule's " +
			"monitor whenever the schedule is active, which is what " +
			"makes the rule recurring instead of one-off.\n\n" +
			"Within a time interval the constraints are combined: a " +
			"moment is inside the interval when it satisfies all of the " +
			"ones that are set, and an omitted constraint matches " +
			"everything. An interval carrying only `weekdays` therefore " +
			"covers those days in full. Across intervals they are " +
			"combined the other way — the schedule is active during any " +
			"of them — so a window that differs by day is expressed as " +
			"one interval per shape.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the muting schedule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "Name of the muting schedule. Must be unique " +
					"across the instance.",
			},
			"time_intervals": schema.ListNestedAttribute{
				Required: true,
				Description: "The recurring windows the schedule is active " +
					"during. At least one is required, and each must set " +
					"at least one constraint.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"times": schema.ListNestedAttribute{
							Optional: true,
							Description: "Spans within the day the interval " +
								"covers. Covers the whole day if unset.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"start_time": schema.StringAttribute{
										Required: true,
										Description: "Inclusive start of the " +
											"span, as 24-hour `HH:MM`.",
									},
									"end_time": schema.StringAttribute{
										Required: true,
										Description: "Exclusive end of the " +
											"span, as 24-hour `HH:MM`. Must " +
											"be after `start_time`, so a " +
											"span crossing midnight is " +
											"written as one interval ending " +
											"at `23:59` and another " +
											"starting at `00:00`.",
									},
								},
							},
						},
						"weekdays": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Days of the week the interval " +
								"covers, as lowercase names such as " +
								"`monday`. Covers every day if unset.",
						},
						"days_of_month": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Days of the month the interval " +
								"covers, each a day or an inclusive " +
								"`begin:end` range such as `1:5`. Days run " +
								"1 to 31, or -31 to -1 counting back from " +
								"the end of the month, so `-1` is the last " +
								"day of whichever month the schedule lands " +
								"in. Covers every day if unset.",
						},
						"months": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Months the interval covers, each " +
								"a lowercase month name such as `january`, " +
								"a number from 1 to 12, or an inclusive " +
								"`begin:end` range of either. Covers every " +
								"month if unset.",
						},
						"years": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Years the interval covers, each a " +
								"year or an inclusive `begin:end` range " +
								"such as `2026:2030`. Covers every year if " +
								"unset.",
						},
						"location": schema.StringAttribute{
							Optional: true,
							Description: "IANA timezone the interval's " +
								"times and days are read in, such as " +
								"`America/New_York`. Defaults to UTC.",
						},
					},
				},
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the muting schedule.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the muting schedule was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "When the muting schedule was last updated.",
			},
		},
	}
}

// ValidateConfig checks the interval fields at plan time. The API
// validates some of them itself, but answers an unknown timezone,
// month or year with a bare 500, which says nothing about which entry
// was wrong.
func (r *mutingScheduleResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var model mutingScheduleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only a fully known list can be judged; one built from another
	// resource's attributes is unknown until apply.
	if model.TimeIntervals.IsNull() || model.TimeIntervals.IsUnknown() {
		return
	}

	intervalsPath := path.Root("time_intervals")
	if len(model.TimeIntervals.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			intervalsPath,
			"Muting schedule has no time intervals",
			"A muting schedule is active during its time intervals, so "+
				"at least one is required.",
		)

		return
	}

	for index, element := range model.TimeIntervals.Elements() {
		validateInterval(
			ctx,
			element,
			intervalsPath.AtListIndex(index),
			&resp.Diagnostics,
		)
	}
}

// validateInterval checks one time interval, reporting each problem
// against the attribute the offending value came from.
func validateInterval(
	ctx context.Context,
	element attr.Value,
	intervalPath path.Path,
	diagnosticsOut *diag.Diagnostics,
) {
	object, ok := element.(types.Object)
	if !ok || object.IsNull() || object.IsUnknown() {
		return
	}

	var interval timeIntervalModel
	diags := object.As(ctx, &interval, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		diagnosticsOut.Append(diags...)
		return
	}

	constraints := []struct {
		name     string
		list     types.List
		validate func(string) error
	}{
		{"weekdays", interval.Weekdays, validateWeekday},
		{"days_of_month", interval.DaysOfMonth, validateDayOfMonth},
		{"months", interval.Months, validateMonth},
		{"years", interval.Years, validateYear},
	}

	isSet := isKnownNonEmpty(interval.Times)
	for _, constraint := range constraints {
		if isKnownNonEmpty(constraint.list) {
			isSet = true
		}

		validateStrings(
			constraint.list,
			constraint.name,
			intervalPath.AtName(constraint.name),
			constraint.validate,
			diagnosticsOut,
		)
	}

	// Every constraint omitted would mute at all times, which the API
	// refuses. A schedule meant to always mute is a muting rule with
	// no schedule and no end.
	if !isSet {
		diagnosticsOut.AddAttributeError(
			intervalPath,
			"Time interval has no constraints",
			"A time interval must set at least one of times, weekdays, "+
				"days_of_month, months, or years. To mute without a "+
				"recurring window, use an oodle_muting_rule with no "+
				"schedule_ids and no ends_at.",
		)
	}

	if isKnown(interval.Location) {
		if err := validateLocation(interval.Location.ValueString()); err != nil {
			diagnosticsOut.AddAttributeError(
				intervalPath.AtName("location"),
				"Invalid timezone",
				err.Error(),
			)
		}
	}

	validateTimes(ctx, interval.Times, intervalPath, diagnosticsOut)
}

// validateTimes checks the spans within the day an interval covers.
func validateTimes(
	ctx context.Context,
	list types.List,
	intervalPath path.Path,
	diagnosticsOut *diag.Diagnostics,
) {
	if !isKnownNonEmpty(list) {
		return
	}

	timesPath := intervalPath.AtName("times")
	for index, element := range list.Elements() {
		object, ok := element.(types.Object)
		if !ok || object.IsNull() || object.IsUnknown() {
			continue
		}

		var timeRange timeRangeModel
		diags := object.As(ctx, &timeRange, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			diagnosticsOut.Append(diags...)
			continue
		}

		timePath := timesPath.AtListIndex(index)
		valid := true
		for _, end := range []struct {
			name  string
			value types.String
		}{
			{"start_time", timeRange.StartTime},
			{"end_time", timeRange.EndTime},
		} {
			if !isKnown(end.value) {
				valid = false
				continue
			}

			if err := validateTimeOfDay(end.value.ValueString()); err != nil {
				diagnosticsOut.AddAttributeError(
					timePath.AtName(end.name),
					"Invalid time of day",
					err.Error(),
				)
				valid = false
			}
		}

		// Comparing malformed times would only add noise to the error
		// already reported for them.
		if !valid {
			continue
		}

		if timeRange.StartTime.ValueString() >= timeRange.EndTime.ValueString() {
			diagnosticsOut.AddAttributeError(
				timePath,
				"Time range ends before it starts",
				"start_time "+timeRange.StartTime.ValueString()+
					" must be before end_time "+
					timeRange.EndTime.ValueString()+
					". A span crossing midnight is written as one range "+
					"ending at \"23:59\" and another starting at \"00:00\".",
			)
		}
	}
}

// validateStrings checks every known entry of one constraint list.
func validateStrings(
	list types.List,
	name string,
	listPath path.Path,
	validate func(string) error,
	diagnosticsOut *diag.Diagnostics,
) {
	if !isKnownNonEmpty(list) {
		return
	}

	for index, element := range list.Elements() {
		value, ok := element.(types.String)
		if !ok || !isKnown(value) {
			continue
		}

		if err := validate(value.ValueString()); err != nil {
			diagnosticsOut.AddAttributeError(
				listPath.AtListIndex(index),
				"Invalid "+name+" entry",
				err.Error(),
			)
		}
	}
}

func isKnown(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func isKnownNonEmpty(list types.List) bool {
	return !list.IsNull() && !list.IsUnknown() && len(list.Elements()) > 0
}
