package monitor

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/validatorutils"
)

func warnCondition() *conditionModel {
	return &conditionModel{
		Operation: types.StringValue(">"),
		Value:     types.Float64Value(1),
		For:       validatorutils.NewDurationValue("5m"),
	}
}

// warn and warning are one threshold: either name sends the
// warning level, and a read keeps the name the config used.
func TestWarnIsTheWarningThreshold(t *testing.T) {
	ctx := context.Background()
	for _, useWarn := range []bool{true, false} {
		m := &monitorResourceModel{Conditions: &conditionsModel{}}
		if useWarn {
			m.Conditions.Warn = warnCondition()
		} else {
			m.Conditions.Warning = warnCondition()
		}
		client := &clientmodels.Monitor{}
		assert.Nil(t, m.ToClientModel(ctx, client))
		assert.NotNil(t, client.Conditions.Warn)
		assert.Equal(t, 1.0, client.Conditions.Warn.Value)

		client.ID = clientmodels.ID{UUID: uuid.New()}
		diags := &diag.Diagnostics{}
		m.FromClientModel(ctx, client, diags)
		assert.False(t, diags.HasError())
		if useWarn {
			assert.NotNil(t, m.Conditions.Warn)
			assert.Nil(t, m.Conditions.Warning)
		} else {
			assert.NotNil(t, m.Conditions.Warning)
			assert.Nil(t, m.Conditions.Warn)
		}
	}
}

func monitorSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	resp := resource.SchemaResponse{}
	(&monitorResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.False(t, resp.Diagnostics.HasError())
	return resp
}

func asObject(t *testing.T, typ tftypes.Type) tftypes.Object {
	t.Helper()
	obj, ok := typ.(tftypes.Object)
	if !ok {
		t.Fatalf("expected tftypes.Object, got %T", typ)
	}
	return obj
}

// monitorConfig builds a config whose conditions set only the
// named thresholds; every other attribute is null.
func monitorConfig(t *testing.T, thresholds ...string) tfsdk.Config {
	t.Helper()
	ctx := context.Background()
	s := monitorSchema(t).Schema
	objType := asObject(t, s.Type().TerraformType(ctx))
	condType := asObject(t, objType.AttributeTypes["conditions"])

	conds := map[string]tftypes.Value{}
	for name, attrType := range condType.AttributeTypes {
		conds[name] = tftypes.NewValue(attrType, nil)
	}
	for _, name := range thresholds {
		levelType := asObject(t, condType.AttributeTypes[name])
		level := map[string]tftypes.Value{}
		for attr, at := range levelType.AttributeTypes {
			level[attr] = tftypes.NewValue(at, nil)
		}
		level["operation"] = tftypes.NewValue(tftypes.String, ">")
		level["value"] = tftypes.NewValue(tftypes.Number, 1)
		level["for"] = tftypes.NewValue(tftypes.String, "5m")
		conds[name] = tftypes.NewValue(levelType, level)
	}

	values := map[string]tftypes.Value{}
	for name, attrType := range objType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["name"] = tftypes.NewValue(tftypes.String, "m")
	values["promql_query"] = tftypes.NewValue(tftypes.String, "up")
	values["conditions"] = tftypes.NewValue(condType, conds)
	return tfsdk.Config{Raw: tftypes.NewValue(objType, values), Schema: s}
}

func validateMonitor(t *testing.T, cfg tfsdk.Config) resource.ValidateConfigResponse {
	t.Helper()
	resp := resource.ValidateConfigResponse{}
	validatorutils.NewMonitorConfigValidator().ValidateResource(
		context.Background(), resource.ValidateConfigRequest{Config: cfg}, &resp,
	)
	return resp
}

func TestConditionsNeedOneThresholdAndOneWarningName(t *testing.T) {
	assert.False(t, validateMonitor(t, monitorConfig(t, "warn")).Diagnostics.HasError(),
		"warn alone is a threshold")
	assert.False(t, validateMonitor(t, monitorConfig(t, "warning", "critical")).Diagnostics.HasError())
	assert.True(t, validateMonitor(t, monitorConfig(t)).Diagnostics.HasError(),
		"a conditions block with no threshold fails at plan time, not on apply")
	assert.True(t, validateMonitor(t, monitorConfig(t, "warning", "warn")).Diagnostics.HasError(),
		"warning and warn together are refused")
}

// In a notifications block, warn and warning are one notifier
// list, and a read keeps the name the config used.
func TestNotificationsWarningIsTheWarnList(t *testing.T) {
	ctx := context.Background()
	notifier := uuid.New()
	ids := types.ListValueMust(types.StringType, []attr.Value{types.StringValue(notifier.String())})
	null := types.ListNull(types.StringType)
	matcherType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"type": types.StringType, "name": types.StringType, "value": types.StringType,
	}}
	itemType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"matchers":               types.ListType{ElemType: matcherType},
		"notification_policy_id": types.StringType,
		"notifiers":              types.ObjectType{AttrTypes: notifiersAttrTypes()},
	}}
	for _, useWarning := range []bool{true, false} {
		n := &notifiersByCondition{Any: null, Warn: null, Warning: null, Critical: null, NoData: null}
		if useWarning {
			n.Warning = ids
		} else {
			n.Warn = ids
		}
		list, diags := types.ListValueFrom(ctx, itemType, []labelMatcherNotificationsModel{{
			Matchers:             types.ListValueMust(matcherType, []attr.Value{}),
			NotificationPolicyID: types.StringNull(),
			Notifiers:            n,
		}})
		assert.False(t, diags.HasError())

		m := &monitorResourceModel{
			Name: types.StringValue("m"), PromQLQuery: types.StringValue("up"),
			Conditions:    &conditionsModel{Critical: warnCondition()},
			Notifications: list,
		}
		client := &clientmodels.Monitor{}
		assert.Nil(t, m.ToClientModel(ctx, client))
		assert.Equal(t, 1, len(client.Notifications))
		assert.Equal(t, notifier, client.Notifications[0].Notifiers.Warn[0].UUID)

		client.ID = clientmodels.ID{UUID: uuid.New()}
		out := &diag.Diagnostics{}
		m.FromClientModel(ctx, client, out)
		assert.False(t, out.HasError())
		var items []labelMatcherNotificationsModel
		assert.False(t, m.Notifications.ElementsAs(ctx, &items, false).HasError())
		assert.Equal(t, 1, len(items))
		if useWarning {
			assert.Equal(t, 1, len(items[0].Notifiers.Warning.Elements()))
			assert.True(t, items[0].Notifiers.Warn.IsNull())
		} else {
			assert.Equal(t, 1, len(items[0].Notifiers.Warn.Elements()))
			assert.True(t, items[0].Notifiers.Warning.IsNull())
		}
	}
}
