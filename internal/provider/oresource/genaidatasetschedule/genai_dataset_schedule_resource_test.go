package genaidatasetschedule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/rubrikinc/testwell/assert"
)

// schemaOf returns the resource schema the validator reads.
func schemaOf(t *testing.T) resource.SchemaResponse {
	t.Helper()
	resp := resource.SchemaResponse{}
	(&genaiDatasetScheduleResource{}).Schema(
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
	schema := schemaOf(t).Schema
	resp := resource.ValidateConfigResponse{}
	(&genaiDatasetScheduleResource{}).ValidateConfig(
		context.Background(),
		resource.ValidateConfigRequest{
			Config: tfsdk.Config{Raw: raw, Schema: schema},
		},
		&resp,
	)

	return resp
}

// configValue builds a config with every attribute null except the
// ones named, so a test states only what it is about.
func configValue(
	t *testing.T,
	set map[string]tftypes.Value,
) tftypes.Value {
	t.Helper()
	objType, ok := schemaOf(t).Schema.Type().
		TerraformType(context.Background()).(tftypes.Object)
	if !ok {
		t.Fatal("resource schema does not describe an object")
	}

	values := map[string]tftypes.Value{}
	for name, attrType := range objType.AttributeTypes {
		if v, ok := set[name]; ok {
			values[name] = v
			continue
		}
		values[name] = tftypes.NewValue(attrType, nil)
	}

	return tftypes.NewValue(objType, values)
}

func strList(values ...string) tftypes.Value {
	listType := tftypes.List{ElementType: tftypes.String}
	elems := make([]tftypes.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, tftypes.NewValue(tftypes.String, v))
	}

	return tftypes.NewValue(listType, elems)
}

// TestValidateConfigSkipsUnknownMode pins that a mode built from
// for_each is left alone. Which fields are required depends on it,
// so judging one branch would report the other branch's fields as
// missing on a config that is perfectly valid.
func TestValidateConfigSkipsUnknownMode(t *testing.T) {
	resp := validate(t, configValue(t, map[string]tftypes.Value{
		"dataset_name": tftypes.NewValue(tftypes.String, "support-eval"),
		"enabled":      tftypes.NewValue(tftypes.Bool, true),
		"mode": tftypes.NewValue(
			tftypes.String, tftypes.UnknownValue,
		),
		"interval_value":    tftypes.NewValue(tftypes.Number, 30),
		"interval_unit":     tftypes.NewValue(tftypes.String, "minutes"),
		"experiment_config": tftypes.NewValue(tftypes.String, "{}"),
	}))
	assert.False(t, resp.Diagnostics.HasError())
}

// TestValidateConfigStillJudgesAKnownMode guards the case above from
// turning into "never validate anything".
func TestValidateConfigStillJudgesAKnownMode(t *testing.T) {
	resp := validate(t, configValue(t, map[string]tftypes.Value{
		"dataset_name":      tftypes.NewValue(tftypes.String, "support-eval"),
		"enabled":           tftypes.NewValue(tftypes.Bool, true),
		"mode":              tftypes.NewValue(tftypes.String, ModeCalendar),
		"interval_value":    tftypes.NewValue(tftypes.Number, 30),
		"interval_unit":     tftypes.NewValue(tftypes.String, "minutes"),
		"times":             strList("09:00"),
		"experiment_config": tftypes.NewValue(tftypes.String, "{}"),
	}))
	assert.True(t, resp.Diagnostics.HasError())
}

// TestHasElementsCountsAnUnknownListAsSet pins that a list built
// from another resource's output is not read as absent.
func TestHasElementsCountsAnUnknownListAsSet(t *testing.T) {
	assert.True(t, hasElements(types.ListUnknown(types.StringType)))
	assert.False(t, hasElements(types.ListNull(types.StringType)))
}
