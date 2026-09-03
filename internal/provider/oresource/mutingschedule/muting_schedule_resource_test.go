package mutingschedule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/rubrikinc/testwell/assert"
)

// schemaOf returns the resource schema the validator reads.
func schemaOf(t *testing.T) resource.SchemaResponse {
	t.Helper()
	resp := resource.SchemaResponse{}
	(&mutingScheduleResource{}).Schema(
		context.Background(),
		resource.SchemaRequest{},
		&resp,
	)
	assert.False(t, resp.Diagnostics.HasError())

	return resp
}

// validate runs ValidateConfig over one config value.
func validate(
	t *testing.T,
	raw tftypes.Value,
) resource.ValidateConfigResponse {
	t.Helper()
	resp := resource.ValidateConfigResponse{}
	(&mutingScheduleResource{}).ValidateConfig(
		context.Background(),
		resource.ValidateConfigRequest{
			Config: tfsdk.Config{Raw: raw, Schema: schemaOf(t).Schema},
		},
		&resp,
	)

	return resp
}

// intervalType is the tftypes shape of one time_intervals element.
func intervalType(t *testing.T) tftypes.Object {
	t.Helper()
	objType, ok := schemaOf(t).Schema.Type().
		TerraformType(context.Background()).(tftypes.Object)
	if !ok {
		t.Fatal("resource schema does not describe an object")
	}

	listType, ok := objType.AttributeTypes["time_intervals"].(tftypes.List)
	if !ok {
		t.Fatal("time_intervals is not a list")
	}

	elemType, ok := listType.ElementType.(tftypes.Object)
	if !ok {
		t.Fatal("time_intervals elements are not objects")
	}

	return elemType
}

// scheduleWith builds a named schedule over the given intervals, so a
// test states only the intervals it is about.
func scheduleWith(t *testing.T, intervals ...tftypes.Value) tftypes.Value {
	t.Helper()
	objType, ok := schemaOf(t).Schema.Type().
		TerraformType(context.Background()).(tftypes.Object)
	if !ok {
		t.Fatal("resource schema does not describe an object")
	}

	values := map[string]tftypes.Value{}
	for name, attrType := range objType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["name"] = tftypes.NewValue(tftypes.String, "schedule")
	values["time_intervals"] = tftypes.NewValue(
		tftypes.List{ElementType: intervalType(t)}, intervals,
	)

	return tftypes.NewValue(objType, values)
}

// interval builds one time interval with only the named attributes
// set.
func interval(t *testing.T, set map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	elemType := intervalType(t)

	values := map[string]tftypes.Value{}
	for name, attrType := range elemType.AttributeTypes {
		if v, ok := set[name]; ok {
			values[name] = v
			continue
		}
		values[name] = tftypes.NewValue(attrType, nil)
	}

	return tftypes.NewValue(elemType, values)
}

func strList(values ...string) tftypes.Value {
	elements := make([]tftypes.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, tftypes.NewValue(tftypes.String, value))
	}

	return tftypes.NewValue(
		tftypes.List{ElementType: tftypes.String}, elements,
	)
}

// times builds a times list from alternating start and end values.
func times(t *testing.T, bounds ...string) tftypes.Value {
	t.Helper()
	rangeType, ok := intervalType(t).AttributeTypes["times"].(tftypes.List)
	if !ok {
		t.Fatal("times is not a list")
	}

	elemType, ok := rangeType.ElementType.(tftypes.Object)
	if !ok {
		t.Fatal("times elements are not objects")
	}

	elements := make([]tftypes.Value, 0, len(bounds)/2)
	for i := 0; i < len(bounds); i += 2 {
		elements = append(elements, tftypes.NewValue(elemType, map[string]tftypes.Value{
			"start_time": tftypes.NewValue(tftypes.String, bounds[i]),
			"end_time":   tftypes.NewValue(tftypes.String, bounds[i+1]),
		}))
	}

	return tftypes.NewValue(rangeType, elements)
}

func TestValidateConfigAcceptsAValidSchedule(t *testing.T) {
	resp := validate(t, scheduleWith(t,
		interval(t, map[string]tftypes.Value{
			"times":         times(t, "09:00", "12:00", "13:00", "17:00"),
			"weekdays":      strList("monday", "friday"),
			"days_of_month": strList("1:5", "-1"),
			"months":        strList("january", "3:6"),
			"years":         strList("2026:2030"),
			"location":      tftypes.NewValue(tftypes.String, "America/New_York"),
		}),
		interval(t, map[string]tftypes.Value{
			"weekdays": strList("saturday", "sunday"),
		}),
	))
	assert.False(
		t, resp.Diagnostics.HasError(), "diagnostics: %v", resp.Diagnostics,
	)
}

func TestValidateConfigRejectsAnEmptyIntervalList(t *testing.T) {
	assert.True(t, validate(t, scheduleWith(t)).Diagnostics.HasError())
}

// TestValidateConfigRejectsAnUnconstrainedInterval covers the
// interval that sets nothing, which would mute at all times.
func TestValidateConfigRejectsAnUnconstrainedInterval(t *testing.T) {
	resp := validate(t, scheduleWith(t,
		interval(t, map[string]tftypes.Value{
			"location": tftypes.NewValue(tftypes.String, "UTC"),
		}),
	))
	assert.True(t, resp.Diagnostics.HasError())
}

// TestValidateConfigCountsTimesAsAConstraint guards the check above
// from rejecting an interval constrained only by time of day.
func TestValidateConfigCountsTimesAsAConstraint(t *testing.T) {
	resp := validate(t, scheduleWith(t,
		interval(t, map[string]tftypes.Value{
			"times": times(t, "22:00", "23:59"),
		}),
	))
	assert.False(
		t, resp.Diagnostics.HasError(), "diagnostics: %v", resp.Diagnostics,
	)
}

func TestValidateConfigRejectsBadConstraintEntries(t *testing.T) {
	for name, value := range map[string]tftypes.Value{
		"weekdays":      strList("monday:friday"),
		"days_of_month": strList("32"),
		"months":        strList("jan"),
		"years":         strList("abcd"),
	} {
		resp := validate(t, scheduleWith(t,
			interval(t, map[string]tftypes.Value{name: value}),
		))
		assert.True(t, resp.Diagnostics.HasError(), "attribute: %v", name)
	}
}

func TestValidateConfigRejectsAnUnknownTimezone(t *testing.T) {
	resp := validate(t, scheduleWith(t,
		interval(t, map[string]tftypes.Value{
			"weekdays": strList("monday"),
			"location": tftypes.NewValue(tftypes.String, "Mars/Olympus"),
		}),
	))
	assert.True(t, resp.Diagnostics.HasError())
}

func TestValidateConfigRejectsAReversedTimeRange(t *testing.T) {
	resp := validate(t, scheduleWith(t,
		interval(t, map[string]tftypes.Value{
			"times": times(t, "17:00", "09:00"),
		}),
	))
	assert.True(t, resp.Diagnostics.HasError())
}

func TestValidateConfigRejectsABadTimeOfDay(t *testing.T) {
	resp := validate(t, scheduleWith(t,
		interval(t, map[string]tftypes.Value{
			"times": times(t, "09:00", "24:00"),
		}),
	))
	assert.True(t, resp.Diagnostics.HasError())
}

// TestValidateConfigSkipsUnknownIntervals pins that intervals built
// from another resource's output are left alone. They are unknown
// until apply, so judging them would fail a config that is fine.
func TestValidateConfigSkipsUnknownIntervals(t *testing.T) {
	objType, ok := schemaOf(t).Schema.Type().
		TerraformType(context.Background()).(tftypes.Object)
	if !ok {
		t.Fatal("resource schema does not describe an object")
	}

	values := map[string]tftypes.Value{}
	for name, attrType := range objType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["name"] = tftypes.NewValue(tftypes.String, "schedule")
	values["time_intervals"] = tftypes.NewValue(
		tftypes.List{ElementType: intervalType(t)}, tftypes.UnknownValue,
	)

	resp := validate(t, tftypes.NewValue(objType, values))
	assert.False(
		t, resp.Diagnostics.HasError(), "diagnostics: %v", resp.Diagnostics,
	)
}

// TestValidateConfigSkipsAnUnknownEntry covers a single constraint
// value that is unknown until apply.
func TestValidateConfigSkipsAnUnknownEntry(t *testing.T) {
	resp := validate(t, scheduleWith(t,
		interval(t, map[string]tftypes.Value{
			"weekdays": tftypes.NewValue(
				tftypes.List{ElementType: tftypes.String},
				[]tftypes.Value{
					tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				},
			),
		}),
	))
	assert.False(
		t, resp.Diagnostics.HasError(), "diagnostics: %v", resp.Diagnostics,
	)
}
