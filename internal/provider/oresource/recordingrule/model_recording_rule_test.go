package recordingrule

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func serverRule() *clientmodels.RecordingRuleEnvelope {
	return &clientmodels.RecordingRuleEnvelope{RecordingRule: &clientmodels.RecordingRule{
		ID:             "5d0c8b1a-3e2f-4a6b-9c7d-8e1f2a3b4c5d",
		Name:           "Checkout request rate",
		MetricName:     "service:http_requests:rate5m",
		PrometheusExpr: "sum by (service) (rate(http_requests_total[5m]))",
		IntervalSecs:   60,
		ExecutionGroup: "5d0c8b1a-3e2f-4a6b-9c7d-8e1f2a3b4c5d",
	}}
}

// minimalPlan is a config that sets only the required attributes.
func minimalPlan() *recordingRuleResourceModel {
	return &recordingRuleResourceModel{
		ID:             types.StringUnknown(),
		Name:           types.StringValue("Checkout request rate"),
		MetricName:     types.StringValue("service:http_requests:rate5m"),
		PrometheusExpr: types.StringValue("sum by (service) (rate(http_requests_total[5m]))\n"),
		IntervalSecs:   types.Int64Null(),
		ExecutionGroup: types.StringNull(),
		Labels:         types.MapNull(types.StringType),
	}
}

func TestCreateWithDefaultsKeepsThePlan(t *testing.T) {
	ctx := context.Background()
	plan := minimalPlan()

	state := &recordingRuleResourceModel{}
	state.TakeNamesFrom(plan)
	var diags diag.Diagnostics
	state.FromClientModel(ctx, serverRule(), &diags)
	assert.False(t, diags.HasError())

	assert.Equal(t, state.ID.ValueString(), "5d0c8b1a-3e2f-4a6b-9c7d-8e1f2a3b4c5d")
	// The server trimmed the heredoc newline. The state keeps the config.
	assert.True(t, state.PrometheusExpr.Equal(plan.PrometheusExpr))
	// The server filled in the defaults. The state keeps them unset.
	assert.True(t, state.IntervalSecs.IsNull())
	assert.True(t, state.ExecutionGroup.IsNull())
	assert.True(t, state.Labels.IsNull())
}

func TestImportReadsEveryValue(t *testing.T) {
	ctx := context.Background()
	rule := serverRule()
	rule.RecordingRule.IntervalSecs = 120
	rule.RecordingRule.ExecutionGroup = "checkout"
	rule.RecordingRule.LabelPolicy = &clientmodels.RecordingRuleLabelPolicy{Add: map[string]string{"team": "checkout"}}

	// An import starts from a state that only has the ID.
	state := &recordingRuleResourceModel{
		ID:             types.StringValue("5d0c8b1a-3e2f-4a6b-9c7d-8e1f2a3b4c5d"),
		IntervalSecs:   types.Int64Null(),
		ExecutionGroup: types.StringNull(),
		Labels:         types.MapNull(types.StringType),
	}
	var diags diag.Diagnostics
	state.FromClientModel(ctx, rule, &diags)
	assert.False(t, diags.HasError())

	assert.Equal(t, state.Name.ValueString(), "Checkout request rate")
	assert.Equal(t, state.IntervalSecs.ValueInt64(), int64(120))
	assert.Equal(t, state.ExecutionGroup.ValueString(), "checkout")
	labels := map[string]string{}
	assert.False(t, state.Labels.ElementsAs(ctx, &labels, false).HasError())
	assert.DeepEqual(t, labels, map[string]string{"team": "checkout"})
}

func TestReadReportsAChangeMadeOutsideTerraform(t *testing.T) {
	ctx := context.Background()
	state := minimalPlan()
	state.ID = types.StringValue("5d0c8b1a-3e2f-4a6b-9c7d-8e1f2a3b4c5d")

	rule := serverRule()
	rule.RecordingRule.PrometheusExpr = "sum(rate(http_requests_total[1m]))"
	rule.RecordingRule.IntervalSecs = 300
	var diags diag.Diagnostics
	state.FromClientModel(ctx, rule, &diags)

	assert.Equal(t, state.PrometheusExpr.ValueString(), "sum(rate(http_requests_total[1m]))")
	assert.Equal(t, state.IntervalSecs.ValueInt64(), int64(300))
}

func TestToClientModelRequestBody(t *testing.T) {
	ctx := context.Background()
	labels, d := types.MapValueFrom(ctx, types.StringType, map[string]string{"team": "checkout"})
	assert.False(t, d.HasError())

	tests := []struct {
		name  string
		model *recordingRuleResourceModel
		want  string
	}{
		{
			name:  "create with defaults",
			model: minimalPlan(),
			want: `{"recording_rule":{"name":"Checkout request rate","metric_name":"service:http_requests:rate5m",` +
				`"prometheus_expr":"sum by (service) (rate(http_requests_total[5m]))\n"}}`,
		},
		{
			name: "update sends the ID from the state",
			model: func() *recordingRuleResourceModel {
				m := minimalPlan()
				m.ID = types.StringValue("5d0c8b1a-3e2f-4a6b-9c7d-8e1f2a3b4c5d")
				m.IntervalSecs = types.Int64Value(120)
				m.ExecutionGroup = types.StringValue("checkout")
				m.Labels = labels
				return m
			}(),
			want: `{"recording_rule":{"id":"5d0c8b1a-3e2f-4a6b-9c7d-8e1f2a3b4c5d","name":"Checkout request rate",` +
				`"metric_name":"service:http_requests:rate5m",` +
				`"prometheus_expr":"sum by (service) (rate(http_requests_total[5m]))\n",` +
				`"interval_secs":120,"execution_group":"checkout","label_policy":{"add":{"team":"checkout"}}}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope := &clientmodels.RecordingRuleEnvelope{}
			assert.Nil(t, tt.model.ToClientModel(ctx, envelope))
			got, err := json.Marshal(envelope)
			assert.Nil(t, err)
			assert.Equal(t, string(got), tt.want)
		})
	}
}
