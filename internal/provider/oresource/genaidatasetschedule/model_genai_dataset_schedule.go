package genaidatasetschedule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type genaiDatasetScheduleResourceModel struct {
	ID               types.String `tfsdk:"id"`
	DatasetName      types.String `tfsdk:"dataset_name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Mode             types.String `tfsdk:"mode"`
	IntervalValue    types.Int64  `tfsdk:"interval_value"`
	IntervalUnit     types.String `tfsdk:"interval_unit"`
	Timezone         types.String `tfsdk:"timezone"`
	Times            types.List   `tfsdk:"times"`
	Weekdays         types.List   `tfsdk:"weekdays"`
	DaysOfMonth      types.List   `tfsdk:"days_of_month"`
	ExperimentConfig types.String `tfsdk:"experiment_config"`
	ScheduleID       types.String `tfsdk:"schedule_id"`
	DatasetID        types.String `tfsdk:"dataset_id"`
	NextRunAt        types.String `tfsdk:"next_run_at"`
	LastRunAt        types.String `tfsdk:"last_run_at"`
	LastError        types.String `tfsdk:"last_error"`
}

func (m *genaiDatasetScheduleResourceModel) GetID() types.String {
	return m.ID
}

func (m *genaiDatasetScheduleResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *genaiDatasetScheduleResourceModel) FromClientModel(
	ctx context.Context,
	model *clientmodels.GenAIDatasetSchedule,
	diagnosticsOut *diag.Diagnostics,
) {
	// The schedule is addressed by dataset name, so that is the id
	// rather than the server-assigned uuid, which appears in no path.
	m.ID = types.StringValue(model.DatasetName)
	m.DatasetName = types.StringValue(model.DatasetName)
	m.ScheduleID = types.StringValue(model.ID)
	m.DatasetID = types.StringValue(model.DatasetID)
	m.Enabled = types.BoolValue(model.Enabled)

	m.NextRunAt = optionalString(model.NextRunAt)
	m.LastRunAt = optionalString(model.LastRunAt)
	m.LastError = optionalString(model.LastError)
	m.ExperimentConfig = resourceutils.RawToJSONString(
		model.ExperimentConfig, m.ExperimentConfig,
	)

	// Empty mode means calendar, the default the server applies.
	m.Mode = takeUnlessDefault(model.Mode, ModeCalendar, m.Mode)

	// A schedule answers with both field groups whatever its mode:
	// the calendar fields carry the defaults the server applied, and
	// the interval fields come back as 0 and "". Only the group the
	// mode actually reads belongs in state — Terraform fails an
	// apply that returns a value for an optional attribute the
	// config left out, and the ignored group is exactly that.
	if model.Mode == ModeInterval {
		m.Timezone = types.StringNull()
		m.Times = types.ListNull(types.StringType)
		m.Weekdays = types.ListNull(types.StringType)
		m.DaysOfMonth = types.ListNull(types.StringType)
		m.IntervalValue = types.Int64Value(model.IntervalValue)
		m.IntervalUnit = optionalString(model.IntervalUnit)

		return
	}

	m.IntervalValue = types.Int64Null()
	m.IntervalUnit = types.StringNull()
	m.Timezone = takeUnlessDefault(
		model.Timezone, defaultTimezone, m.Timezone,
	)
	m.Times = resourceutils.SliceToStringList(
		ctx, model.Times, m.Times, diagnosticsOut,
	)
	m.Weekdays = resourceutils.SliceToStringList(
		ctx, model.Weekdays, m.Weekdays, diagnosticsOut,
	)
	m.DaysOfMonth = resourceutils.SliceToStringList(
		ctx, model.DaysOfMonth, m.DaysOfMonth, diagnosticsOut,
	)
}

func (m *genaiDatasetScheduleResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.GenAIDatasetSchedule,
) error {
	model.DatasetName = m.DatasetName.ValueString()
	model.Enabled = m.Enabled.ValueBool()
	model.Mode = m.Mode.ValueString()
	model.IntervalValue = m.IntervalValue.ValueInt64()
	model.IntervalUnit = m.IntervalUnit.ValueString()
	model.Timezone = m.Timezone.ValueString()

	times, err := resourceutils.StringListToSlice(ctx, m.Times)
	if err != nil {
		return err
	}
	model.Times = times

	weekdays, err := resourceutils.StringListToSlice(ctx, m.Weekdays)
	if err != nil {
		return err
	}
	model.Weekdays = weekdays

	daysOfMonth, err := resourceutils.StringListToSlice(ctx, m.DaysOfMonth)
	if err != nil {
		return err
	}
	model.DaysOfMonth = daysOfMonth

	experimentConfig, err := resourceutils.JSONStringToRaw(
		m.ExperimentConfig, "experiment_config",
	)
	if err != nil {
		return err
	}
	model.ExperimentConfig = experimentConfig

	return nil
}

// takeUnlessDefault keeps an attribute null when the server merely
// echoed the default it applies to an unset field.
//
// Terraform rejects an apply whose result gives an optional
// attribute a value the config did not ask for, so an echoed default
// has to be recognised as one rather than written into state.
func takeUnlessDefault(
	value, serverDefault string,
	prior types.String,
) types.String {
	if prior.IsNull() && value == serverDefault {
		return types.StringNull()
	}

	return optionalString(value)
}

// optionalString keeps an unset optional attribute null rather than
// turning it into "" when the API omits the field.
func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}
