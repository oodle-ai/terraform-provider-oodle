package genaidatasetschedule

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// intervalSchedule is what the API answers for an interval schedule:
// the interval fields it reads, plus the calendar fields it does not,
// carrying the defaults it stored anyway.
func intervalSchedule() *clientmodels.GenAIDatasetSchedule {
	return &clientmodels.GenAIDatasetSchedule{
		DatasetName:      "support-eval",
		ID:               "schedule-uuid",
		DatasetID:        "dataset-uuid",
		Enabled:          true,
		Mode:             ModeInterval,
		IntervalValue:    6,
		IntervalUnit:     "hours",
		Timezone:         "UTC",
		Times:            []string{},
		Weekdays:         []string{},
		DaysOfMonth:      []string{},
		ExperimentConfig: json.RawMessage(`{"datasetId":"dataset-uuid"}`),
		NextRunAt:        "2026-08-20T06:00:00Z",
	}
}

func TestScheduleRoundTrip(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.GenAIDatasetSchedule{
		DatasetName:   "support-eval",
		ID:            "schedule-uuid",
		DatasetID:     "dataset-uuid",
		Enabled:       true,
		Mode:          ModeCalendar,
		Timezone:      "America/Los_Angeles",
		Times:         []string{"09:00", "21:30"},
		Weekdays:      []string{"monday", "friday"},
		DaysOfMonth:   []string{"1"},
		IntervalValue: 0,
		ExperimentConfig: json.RawMessage(
			`{"datasetId":"dataset-uuid","promptName":"support-reply"}`,
		),
	}

	resourceModel := &genaiDatasetScheduleResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	// The schedule is addressed by dataset name, so that is the id.
	assert.Equal(t, "support-eval", resourceModel.ID.ValueString())
	assert.Equal(t, "schedule-uuid", resourceModel.ScheduleID.ValueString())

	written := &clientmodels.GenAIDatasetSchedule{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, written))
	assert.DeepEqual(t, clientModel.Times, written.Times)
	assert.DeepEqual(t, clientModel.Weekdays, written.Weekdays)
	assert.DeepEqual(t, clientModel.DaysOfMonth, written.DaysOfMonth)
	assert.Equal(t, clientModel.Timezone, written.Timezone)
	assert.Equal(t, clientModel.Enabled, written.Enabled)
	// Nothing configured mode, so the echoed "calendar" was the
	// server's default and is written back as empty — which the
	// server reads as calendar again.
	assert.Equal(t, "", written.Mode)
}

// TestScheduleDropsTheOtherModesFields pins that only the fields the
// schedule's mode reads reach state. Terraform fails an apply whose
// result gives an optional attribute a value the config left out, and
// the ignored group is exactly that: an interval schedule is stored
// with timezone "UTC" whether or not anyone asked for one.
func TestScheduleDropsTheOtherModesFields(t *testing.T) {
	ctx := context.Background()
	resourceModel := &genaiDatasetScheduleResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, intervalSchedule(), diags)
	assert.False(t, diags.HasError())

	assert.True(t, resourceModel.Timezone.IsNull())
	assert.True(t, resourceModel.Times.IsNull())
	assert.True(t, resourceModel.Weekdays.IsNull())
	assert.True(t, resourceModel.DaysOfMonth.IsNull())
	assert.Equal(t, int64(6), resourceModel.IntervalValue.ValueInt64())
	assert.Equal(t, "hours", resourceModel.IntervalUnit.ValueString())

	// And the other way round: a calendar schedule answers with the
	// interval fields zeroed, which is not a configured 0.
	calendar := intervalSchedule()
	calendar.Mode = ModeCalendar
	calendar.IntervalValue = 0
	calendar.IntervalUnit = ""
	calendar.Times = []string{"09:00"}

	resourceModel = &genaiDatasetScheduleResourceModel{}
	resourceModel.FromClientModel(ctx, calendar, diags)
	assert.False(t, diags.HasError())
	assert.True(t, resourceModel.IntervalValue.IsNull())
	assert.True(t, resourceModel.IntervalUnit.IsNull())
}

// TestScheduleKeepsServerDefaultsNull pins that a default the server
// applied to a field nobody set stays out of state, and that an
// explicitly configured value of the same string does not.
func TestScheduleKeepsServerDefaultsNull(t *testing.T) {
	ctx := context.Background()
	schedule := intervalSchedule()
	schedule.Mode = ModeCalendar
	schedule.IntervalValue = 0
	schedule.IntervalUnit = ""
	schedule.Times = []string{"09:00"}

	// Nothing was configured, so "calendar" and "UTC" are the
	// server's own defaults rather than answers to the config.
	resourceModel := &genaiDatasetScheduleResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, schedule, diags)
	assert.False(t, diags.HasError())
	assert.True(t, resourceModel.Mode.IsNull())
	assert.True(t, resourceModel.Timezone.IsNull())

	// Configured to the same values, they are kept: dropping them
	// would report a change the config did ask for.
	resourceModel = &genaiDatasetScheduleResourceModel{
		Mode:     types.StringValue(ModeCalendar),
		Timezone: types.StringValue(defaultTimezone),
	}
	resourceModel.FromClientModel(ctx, schedule, diags)
	assert.False(t, diags.HasError())
	assert.Equal(t, ModeCalendar, resourceModel.Mode.ValueString())
	assert.Equal(t, defaultTimezone, resourceModel.Timezone.ValueString())
}
