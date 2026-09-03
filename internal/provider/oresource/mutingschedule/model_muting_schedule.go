package mutingschedule

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

var timeRangeAttrTypes = map[string]attr.Type{
	"start_time": types.StringType,
	"end_time":   types.StringType,
}

var timeIntervalAttrTypes = map[string]attr.Type{
	"times": types.ListType{
		ElemType: types.ObjectType{AttrTypes: timeRangeAttrTypes},
	},
	"weekdays":      types.ListType{ElemType: types.StringType},
	"days_of_month": types.ListType{ElemType: types.StringType},
	"months":        types.ListType{ElemType: types.StringType},
	"years":         types.ListType{ElemType: types.StringType},
	"location":      types.StringType,
}

type timeRangeModel struct {
	StartTime types.String `tfsdk:"start_time"`
	EndTime   types.String `tfsdk:"end_time"`
}

type timeIntervalModel struct {
	Times       types.List   `tfsdk:"times"`
	Weekdays    types.List   `tfsdk:"weekdays"`
	DaysOfMonth types.List   `tfsdk:"days_of_month"`
	Months      types.List   `tfsdk:"months"`
	Years       types.List   `tfsdk:"years"`
	Location    types.String `tfsdk:"location"`
}

type mutingScheduleResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	TimeIntervals types.List   `tfsdk:"time_intervals"`
	CreatedBy     types.String `tfsdk:"created_by"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func (m *mutingScheduleResourceModel) GetID() types.String {
	return m.ID
}

func (m *mutingScheduleResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *mutingScheduleResourceModel) FromClientModel(
	_ context.Context,
	model *clientmodels.MutingSchedule,
	diagnosticsOut *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.CreatedBy = optionalString(model.CreatedBy)
	m.CreatedAt = optionalString(model.CreatedAt)
	m.UpdatedAt = optionalString(model.UpdatedAt)
	m.TimeIntervals = intervalsFromClient(model.TimeIntervals, diagnosticsOut)
}

func (m *mutingScheduleResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.MutingSchedule,
) error {
	model.ID = m.ID.ValueString()
	model.Name = m.Name.ValueString()

	intervals, err := intervalsToClient(ctx, m.TimeIntervals)
	if err != nil {
		return err
	}
	model.TimeIntervals = intervals

	return nil
}

// intervalsToClient converts the configured time intervals into the
// client model.
func intervalsToClient(
	ctx context.Context,
	list types.List,
) ([]clientmodels.ScheduleTimeInterval, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	intervals := make(
		[]clientmodels.ScheduleTimeInterval, 0, len(list.Elements()),
	)
	for _, element := range list.Elements() {
		object, ok := element.(types.Object)
		if !ok {
			return nil, fmt.Errorf(
				"failed to parse time interval: %v, type is %T",
				element,
				element,
			)
		}

		var interval timeIntervalModel
		diags := object.As(ctx, &interval, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf(
				"failed to parse time interval fields: %v", diags,
			)
		}

		times, err := timesToClient(ctx, interval.Times)
		if err != nil {
			return nil, err
		}

		weekdays, err := stringsToClient(ctx, interval.Weekdays, "weekdays")
		if err != nil {
			return nil, err
		}

		daysOfMonth, err := stringsToClient(
			ctx, interval.DaysOfMonth, "days_of_month",
		)
		if err != nil {
			return nil, err
		}

		months, err := stringsToClient(ctx, interval.Months, "months")
		if err != nil {
			return nil, err
		}

		years, err := stringsToClient(ctx, interval.Years, "years")
		if err != nil {
			return nil, err
		}

		intervals = append(intervals, clientmodels.ScheduleTimeInterval{
			Times:       times,
			Weekdays:    weekdays,
			DaysOfMonth: daysOfMonth,
			Months:      months,
			Years:       years,
			Location:    interval.Location.ValueString(),
		})
	}

	return intervals, nil
}

func timesToClient(
	ctx context.Context,
	list types.List,
) ([]clientmodels.ScheduleTimeRange, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	times := make([]clientmodels.ScheduleTimeRange, 0, len(list.Elements()))
	for _, element := range list.Elements() {
		object, ok := element.(types.Object)
		if !ok {
			return nil, fmt.Errorf(
				"failed to parse time range: %v, type is %T",
				element,
				element,
			)
		}

		var timeRange timeRangeModel
		diags := object.As(ctx, &timeRange, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, fmt.Errorf(
				"failed to parse time range fields: %v", diags,
			)
		}

		times = append(times, clientmodels.ScheduleTimeRange{
			StartTime: timeRange.StartTime.ValueString(),
			EndTime:   timeRange.EndTime.ValueString(),
		})
	}

	return times, nil
}

func stringsToClient(
	ctx context.Context,
	list types.List,
	field string,
) ([]string, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var values []string
	diags := list.ElementsAs(ctx, &values, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse %s: %v", field, diags)
	}

	return values, nil
}

// intervalsFromClient converts the API's time intervals back into
// state.
func intervalsFromClient(
	intervals []clientmodels.ScheduleTimeInterval,
	diagnosticsOut *diag.Diagnostics,
) types.List {
	intervalType := types.ObjectType{AttrTypes: timeIntervalAttrTypes}

	elements := make([]attr.Value, 0, len(intervals))
	for _, interval := range intervals {
		object, diags := types.ObjectValue(
			timeIntervalAttrTypes,
			map[string]attr.Value{
				"times":         timesFromClient(interval.Times, diagnosticsOut),
				"weekdays":      stringsFromClient(interval.Weekdays),
				"days_of_month": stringsFromClient(interval.DaysOfMonth),
				"months":        stringsFromClient(interval.Months),
				"years":         stringsFromClient(interval.Years),
				"location":      optionalString(interval.Location),
			},
		)
		if diags.HasError() {
			diagnosticsOut.Append(diags...)
			continue
		}

		elements = append(elements, object)
	}

	list, diags := types.ListValue(intervalType, elements)
	if diags.HasError() {
		diagnosticsOut.Append(diags...)
		return types.ListNull(intervalType)
	}

	return list
}

func timesFromClient(
	times []clientmodels.ScheduleTimeRange,
	diagnosticsOut *diag.Diagnostics,
) types.List {
	timeRangeType := types.ObjectType{AttrTypes: timeRangeAttrTypes}
	if len(times) == 0 {
		return types.ListNull(timeRangeType)
	}

	elements := make([]attr.Value, 0, len(times))
	for _, timeRange := range times {
		object, diags := types.ObjectValue(
			timeRangeAttrTypes,
			map[string]attr.Value{
				"start_time": types.StringValue(timeRange.StartTime),
				"end_time":   types.StringValue(timeRange.EndTime),
			},
		)
		if diags.HasError() {
			diagnosticsOut.Append(diags...)
			continue
		}

		elements = append(elements, object)
	}

	list, diags := types.ListValue(timeRangeType, elements)
	if diags.HasError() {
		diagnosticsOut.Append(diags...)
		return types.ListNull(timeRangeType)
	}

	return list
}

// stringsFromClient keeps an omitted constraint null rather than
// turning it into an empty list, which Terraform would report as a
// change against a config that left it out.
func stringsFromClient(values []string) types.List {
	if len(values) == 0 {
		return types.ListNull(types.StringType)
	}

	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.StringValue(value))
	}

	return types.ListValueMust(types.StringType, elements)
}

// optionalString keeps an unset optional attribute null rather than
// turning it into "" when the API omits the field.
func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}
