package mutingschedule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// TestRoundTrip pins that a schedule read back from the API converts
// into the same client model it came from, so a plan over an
// unchanged schedule is empty.
func TestRoundTrip(t *testing.T) {
	ctx := context.Background()
	original := &clientmodels.MutingSchedule{
		ID:   "3499830c-83d5-4c9d-9f13-706da51cba3f",
		Name: "weekday-business-hours",
		TimeIntervals: []clientmodels.ScheduleTimeInterval{
			{
				Times: []clientmodels.ScheduleTimeRange{
					{StartTime: "09:00", EndTime: "12:00"},
					{StartTime: "13:00", EndTime: "17:00"},
				},
				Weekdays:    []string{"monday", "friday"},
				DaysOfMonth: []string{"1:5", "-1"},
				Months:      []string{"january", "3:6"},
				Years:       []string{"2026:2030"},
				Location:    "America/New_York",
			},
			{
				Weekdays: []string{"saturday"},
			},
		},
		CreatedBy: "team@oodle.ai",
		CreatedAt: "2026-03-13T09:11:04.750631Z",
		UpdatedAt: "2026-03-13T09:11:04.750631Z",
	}

	var diagnostics diag.Diagnostics
	model := &mutingScheduleResourceModel{}
	model.FromClientModel(ctx, original, &diagnostics)
	assert.False(t, diagnostics.HasError(), "diagnostics: %v", diagnostics)

	roundTripped := &clientmodels.MutingSchedule{}
	assert.Nil(t, model.ToClientModel(ctx, roundTripped))
	assert.Equal(t, original.ID, roundTripped.ID)
	assert.Equal(t, original.Name, roundTripped.Name)
	assert.DeepEqual(t, original.TimeIntervals, roundTripped.TimeIntervals)
}

// TestOmittedConstraintsStayNull pins that a constraint the config
// left out reads back as null rather than as an empty list, which
// Terraform would report as a change on every plan.
func TestOmittedConstraintsStayNull(t *testing.T) {
	ctx := context.Background()
	var diagnostics diag.Diagnostics
	model := &mutingScheduleResourceModel{}
	model.FromClientModel(ctx, &clientmodels.MutingSchedule{
		ID:   "id",
		Name: "weekends",
		TimeIntervals: []clientmodels.ScheduleTimeInterval{
			{Weekdays: []string{"saturday", "sunday"}},
		},
	}, &diagnostics)
	assert.False(t, diagnostics.HasError(), "diagnostics: %v", diagnostics)

	interval, ok := model.TimeIntervals.Elements()[0].(types.Object)
	assert.True(t, ok)

	attributes := interval.Attributes()
	for _, name := range []string{
		"times", "days_of_month", "months", "years",
	} {
		assert.True(
			t, attributes[name].IsNull(), "attribute: %v", name,
		)
	}

	// location is unset on the server too, and must not become "".
	assert.True(t, attributes["location"].IsNull())
	assert.False(t, attributes["weekdays"].IsNull())

	// created_by is absent for a schedule the API records no author
	// for, and must not become "" either.
	assert.True(t, model.CreatedBy.IsNull())
}
