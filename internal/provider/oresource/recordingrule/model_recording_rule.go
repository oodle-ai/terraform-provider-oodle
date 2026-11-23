package recordingrule

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

// The server runs a rule at this interval when the request has none.
const defaultIntervalSecs = 60

type recordingRuleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Slug           types.String `tfsdk:"slug"`
	Name           types.String `tfsdk:"name"`
	MetricName     types.String `tfsdk:"metric_name"`
	PrometheusExpr types.String `tfsdk:"prometheus_expr"`
	IntervalSecs   types.Int64  `tfsdk:"interval_secs"`
	ExecutionGroup types.String `tfsdk:"execution_group"`
	Labels         types.Map    `tfsdk:"labels"`
}

var (
	_ resourceutils.ResourceModel[*clientmodels.RecordingRuleEnvelope] = (*recordingRuleResourceModel)(nil)
	_ resourceutils.NameKeeper[*recordingRuleResourceModel]            = (*recordingRuleResourceModel)(nil)
)

func (m *recordingRuleResourceModel) GetID() types.String {
	return m.ID
}

func (m *recordingRuleResourceModel) SetID(id types.String) {
	m.ID = id
}

// TakeNamesFrom copies the planned values before a create or an update
// reads the API answer. The server trims text and fills in defaults, so
// FromClientModel compares the answer with the plan: an answer that only
// differs that way keeps the planned value. Without it, Terraform reports
// an inconsistent result or plans a change on each run.
func (m *recordingRuleResourceModel) TakeNamesFrom(plan *recordingRuleResourceModel) {
	*m = *plan
}

func (m *recordingRuleResourceModel) FromClientModel(
	ctx context.Context,
	envelope *clientmodels.RecordingRuleEnvelope,
	diagnosticsOut *diag.Diagnostics,
) {
	if envelope.RecordingRule == nil {
		diagnosticsOut.AddError("Empty response", "The API response has no recording_rule.")
		return
	}
	rule := envelope.RecordingRule
	prior := *m

	m.ID = types.StringValue(rule.Slug)
	m.Slug = types.StringValue(rule.Slug)
	m.Name = keepUntrimmed(prior.Name, rule.Name)
	m.MetricName = keepUntrimmed(prior.MetricName, rule.MetricName)
	m.PrometheusExpr = keepUntrimmed(prior.PrometheusExpr, rule.PrometheusExpr)

	// An unset value stays null when the server gave it the default, so
	// a config that does not set it has no diff.
	m.IntervalSecs = types.Int64Value(rule.IntervalSecs)
	if prior.IntervalSecs.IsNull() && rule.IntervalSecs == defaultIntervalSecs {
		m.IntervalSecs = types.Int64Null()
	}
	m.ExecutionGroup = types.StringValue(rule.ExecutionGroup)
	if prior.ExecutionGroup.IsNull() && rule.ExecutionGroup == rule.Slug {
		m.ExecutionGroup = types.StringNull()
	}

	var labels map[string]string
	if rule.LabelPolicy != nil {
		labels = rule.LabelPolicy.Add
	}
	switch {
	case len(labels) > 0:
		v, d := types.MapValueFrom(ctx, types.StringType, labels)
		diagnosticsOut.Append(d...)
		m.Labels = v
	case !prior.Labels.IsNull() && !prior.Labels.IsUnknown() && len(prior.Labels.Elements()) == 0:
		// An empty map in the config stays empty.
		m.Labels = prior.Labels
	default:
		m.Labels = types.MapNull(types.StringType)
	}
}

func (m *recordingRuleResourceModel) ToClientModel(
	ctx context.Context,
	envelope *clientmodels.RecordingRuleEnvelope,
) error {
	rule := &clientmodels.RecordingRule{
		Name:           m.Name.ValueString(),
		MetricName:     m.MetricName.ValueString(),
		PrometheusExpr: m.PrometheusExpr.ValueString(),
		IntervalSecs:   m.IntervalSecs.ValueInt64(),
		ExecutionGroup: m.ExecutionGroup.ValueString(),
	}
	// On update, the ID holds the slug from the state.
	switch {
	case !m.ID.IsNull() && !m.ID.IsUnknown():
		rule.Slug = m.ID.ValueString()
	case !m.Slug.IsNull() && !m.Slug.IsUnknown():
		rule.Slug = m.Slug.ValueString()
	}

	if !m.Labels.IsNull() && !m.Labels.IsUnknown() {
		labels := map[string]string{}
		if d := m.Labels.ElementsAs(ctx, &labels, false); d.HasError() {
			return diagError(d)
		}
		if len(labels) > 0 {
			rule.LabelPolicy = &clientmodels.RecordingRuleLabelPolicy{Add: labels}
		}
	}

	envelope.RecordingRule = rule
	return nil
}

// keepUntrimmed keeps the prior value when the server only trimmed it,
// for example the last newline of a heredoc.
func keepUntrimmed(prior types.String, got string) types.String {
	if !prior.IsNull() && !prior.IsUnknown() && strings.TrimSpace(prior.ValueString()) == got {
		return prior
	}
	return types.StringValue(got)
}

type diagErr struct{ d diag.Diagnostics }

func (e diagErr) Error() string {
	var msgs []string
	for _, x := range e.d.Errors() {
		msgs = append(msgs, x.Summary()+": "+x.Detail())
	}
	return strings.Join(msgs, "; ")
}

func diagError(d diag.Diagnostics) error {
	return diagErr{d}
}
