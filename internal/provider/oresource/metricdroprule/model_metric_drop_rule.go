package metricdroprule

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	amlabels "github.com/prometheus/alertmanager/pkg/labels"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type metricDropRuleResourceModel struct {
	ID         types.String         `tfsdk:"id"`
	RuleName   types.String         `tfsdk:"rule_name"`
	Type       types.String         `tfsdk:"type"`
	MetricName *dropRuleMatcherModel `tfsdk:"metric_name"`
	Filters    []dropRuleMatcherModel `tfsdk:"filters"`
}

type dropRuleMatcherModel struct {
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

var _ resourceutils.ResourceModel[*clientmodels.MetricDropRule] = (*metricDropRuleResourceModel)(nil)

func (m *metricDropRuleResourceModel) GetID() types.String {
	return m.ID
}

func (m *metricDropRuleResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *metricDropRuleResourceModel) FromClientModel(
	_ context.Context,
	model *clientmodels.MetricDropRule,
	_ *diag.Diagnostics,
) {
	// Reset the model to clear any existing data.
	*m = metricDropRuleResourceModel{}

	m.ID = types.StringValue(model.ID)
	m.RuleName = types.StringValue(model.RuleName)
	m.Type = types.StringValue(model.Type)

	if model.MetricName != nil {
		m.MetricName = &dropRuleMatcherModel{
			Name:  types.StringValue(model.MetricName.Name),
			Type:  types.StringValue(model.MetricName.Type.String()),
			Value: types.StringValue(model.MetricName.Value),
		}
	}

	if len(model.Filters) > 0 {
		m.Filters = make([]dropRuleMatcherModel, len(model.Filters))
		for i, filter := range model.Filters {
			m.Filters[i] = dropRuleMatcherModel{
				Name:  types.StringValue(filter.Name),
				Type:  types.StringValue(filter.Type.String()),
				Value: types.StringValue(filter.Value),
			}
		}
	}
}

func (m *metricDropRuleResourceModel) ToClientModel(
	_ context.Context,
	model *clientmodels.MetricDropRule,
) error {
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		model.ID = m.ID.ValueString()
	}

	model.RuleName = m.RuleName.ValueString()
	model.Type = m.Type.ValueString()

	if m.MetricName != nil {
		matchType, err := parseMatchType(m.MetricName.Type.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse metric_name match type: %v", err)
		}
		model.MetricName = &clientmodels.LabelMatcher{
			Name:  m.MetricName.Name.ValueString(),
			Type:  matchType,
			Value: m.MetricName.Value.ValueString(),
		}
	}

	if len(m.Filters) > 0 {
		model.Filters = make([]*clientmodels.LabelMatcher, len(m.Filters))
		for i, filter := range m.Filters {
			matchType, err := parseMatchType(filter.Type.ValueString())
			if err != nil {
				return fmt.Errorf("failed to parse filter match type: %v", err)
			}
			model.Filters[i] = &clientmodels.LabelMatcher{
				Name:  filter.Name.ValueString(),
				Type:  matchType,
				Value: filter.Value.ValueString(),
			}
		}
	}

	return nil
}

func parseMatchType(s string) (amlabels.MatchType, error) {
	switch s {
	case "=":
		return amlabels.MatchEqual, nil
	case "!=":
		return amlabels.MatchNotEqual, nil
	case "=~":
		return amlabels.MatchRegexp, nil
	case "!~":
		return amlabels.MatchNotRegexp, nil
	default:
		return 0, fmt.Errorf("invalid match type: %s", s)
	}
}
