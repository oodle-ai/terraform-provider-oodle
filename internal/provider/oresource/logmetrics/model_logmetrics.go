package logmetrics

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type logMetricsResourceModel struct {
	ID                types.String            `tfsdk:"id"`
	Name              types.String            `tfsdk:"name"`
	Labels            []labelModel            `tfsdk:"labels"`
	Filter            *filterModel            `tfsdk:"filter"`
	MetricDefinitions []metricDefinitionModel `tfsdk:"metric_definitions"`
}

type labelModel struct {
	Name           types.String         `tfsdk:"name"`
	Value          types.String         `tfsdk:"value"`
	ValueExtractor *valueExtractorModel `tfsdk:"value_extractor"`
}

type valueExtractorModel struct {
	Field    types.String `tfsdk:"field"`
	JSONPath types.String `tfsdk:"json_path"`
	Regex    types.String `tfsdk:"regex"`
}

type matchModel struct {
	Field    types.String `tfsdk:"field"`
	JSONPath types.String `tfsdk:"json_path"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
}

type allNestedFilterModel struct {
	Match *matchModel           `tfsdk:"match"`
	Not   *notNestedFilterModel `tfsdk:"not"`
}

type anyNestedFilterModel struct {
	Match *matchModel            `tfsdk:"match"`
	Not   *notNestedFilterModel  `tfsdk:"not"`
	All   []allNestedFilterModel `tfsdk:"all"`
}

type notNestedFilterModel struct {
	Match *matchModel `tfsdk:"match"`
}

type filterModel struct {
	Match *matchModel            `tfsdk:"match"`
	All   []allNestedFilterModel `tfsdk:"all"`
	Any   []anyNestedFilterModel `tfsdk:"any"`
	Not   *notNestedFilterModel  `tfsdk:"not"`
}

type metricDefinitionModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Field    types.String `tfsdk:"field"`
	JSONPath types.String `tfsdk:"json_path"`
	Regex    types.String `tfsdk:"regex"`
}

var _ resourceutils.ResourceModel[*clientmodels.LogMetrics] = (*logMetricsResourceModel)(nil)

func (m *logMetricsResourceModel) GetID() types.String {
	return m.ID
}

func (m *logMetricsResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *logMetricsResourceModel) FromClientModel(
	ctx context.Context,
	model *clientmodels.LogMetrics,
	diagnosticsOut *diag.Diagnostics,
) {
	// Reset the model to clear any existing data
	*m = logMetricsResourceModel{}

	m.ID = types.StringValue(model.ID.UUID.String())
	m.Name = types.StringValue(model.Name)

	// Convert labels
	if len(model.Labels) > 0 {
		m.Labels = make([]labelModel, len(model.Labels))
		for i, label := range model.Labels {
			m.Labels[i] = labelModel{
				Name: types.StringValue(label.Name),
			}
			if label.Value != nil {
				m.Labels[i].Value = types.StringValue(*label.Value)
			}
			if label.ValueExtractor != nil {
				m.Labels[i].ValueExtractor = &valueExtractorModel{}
				if label.ValueExtractor.Field != nil {
					m.Labels[i].ValueExtractor.Field = types.StringValue(*label.ValueExtractor.Field)
				}
				if label.ValueExtractor.JSONPath != nil {
					m.Labels[i].ValueExtractor.JSONPath = types.StringValue(*label.ValueExtractor.JSONPath)
				}
				if label.ValueExtractor.Regex != nil {
					m.Labels[i].ValueExtractor.Regex = types.StringValue(*label.ValueExtractor.Regex)
				}
			}
		}
	}

	// Convert filter
	if model.Filter != nil {
		m.Filter = &filterModel{}
		if model.Filter.Match != nil {
			m.Filter.Match = m.fromClientModelMatchModel(model.Filter.Match)
		}
		if model.Filter.MatchAll != nil && len(model.Filter.MatchAll.All) > 0 {
			m.Filter.All = make([]allNestedFilterModel, len(model.Filter.MatchAll.All))
			for i, allElem := range model.Filter.MatchAll.All {
				if allElem.Match != nil {
					m.Filter.All[i] = allNestedFilterModel{
						Match: m.fromClientModelMatchModel(allElem.Match),
					}
				}
				if allElem.MatchNone != nil && allElem.MatchNone.Not != nil && allElem.MatchNone.Not.Match != nil {
					m.Filter.All[i] = allNestedFilterModel{
						Not: &notNestedFilterModel{
							Match: m.fromClientModelMatchModel(allElem.MatchNone.Not.Match),
						},
					}
				}
			}
		}
		if model.Filter.MatchAny != nil && len(model.Filter.MatchAny.Any) > 0 {
			m.Filter.Any = make([]anyNestedFilterModel, len(model.Filter.MatchAny.Any))
			for i, anyElem := range model.Filter.MatchAny.Any {
				if anyElem.Match != nil {
					m.Filter.Any[i] = anyNestedFilterModel{
						Match: m.fromClientModelMatchModel(anyElem.Match),
					}
				}
				if anyElem.MatchNone != nil && anyElem.MatchNone.Not != nil && anyElem.Not.Match != nil {
					m.Filter.Any[i] = anyNestedFilterModel{
						Not: &notNestedFilterModel{
							Match: m.fromClientModelMatchModel(anyElem.Not.Match),
						},
					}
				}
				if anyElem.MatchAll != nil && anyElem.MatchAll.All != nil && len(anyElem.MatchAll.All) > 0 {
					m.Filter.Any[i] = anyNestedFilterModel{
						All: make([]allNestedFilterModel, len(anyElem.All)),
					}
					for j, allElem := range anyElem.All {
						m.Filter.Any[i].All[j] = allNestedFilterModel{
							Match: m.fromClientModelMatchModel(allElem.Match),
						}
					}
				}
			}
		}
		if model.Filter.MatchNone != nil && model.Filter.MatchNone.Not != nil && model.Filter.MatchNone.Not.Match != nil {
			m.Filter.Not = &notNestedFilterModel{
				Match: m.fromClientModelMatchModel(model.Filter.MatchNone.Not.Match),
			}
		}
	}

	// Convert metric definitions
	if len(model.MetricDefinitions) > 0 {
		m.MetricDefinitions = make([]metricDefinitionModel, len(model.MetricDefinitions))
		for i, def := range model.MetricDefinitions {
			m.MetricDefinitions[i] = metricDefinitionModel{
				Name: types.StringValue(def.Name),
				Type: types.StringValue(string(def.Type)),
			}
			if def.Field != "" {
				m.MetricDefinitions[i].Field = types.StringValue(def.Field)
			}
			if def.JSONPath != nil {
				m.MetricDefinitions[i].JSONPath = types.StringValue(*def.JSONPath)
			}
			if def.Regex != nil {
				m.MetricDefinitions[i].Regex = types.StringValue(*def.Regex)
			}
		}
	}
}

func (m *logMetricsResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.LogMetrics,
) error {
	var err error
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		model.ID.UUID, err = uuid.Parse(m.ID.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse ID UUID %v: %v", m.ID.ValueString(), err)
		}
	}

	model.Name = m.Name.ValueString()

	// Convert labels
	if len(m.Labels) > 0 {
		model.Labels = make([]*clientmodels.Label, len(m.Labels))
		for i, label := range m.Labels {
			model.Labels[i] = &clientmodels.Label{
				Name: label.Name.ValueString(),
			}
			if !label.Value.IsNull() {
				value := label.Value.ValueString()
				model.Labels[i].Value = &value
			}
			if label.ValueExtractor != nil {
				model.Labels[i].ValueExtractor = &clientmodels.ValueExtractor{}
				if !label.ValueExtractor.Field.IsNull() {
					field := label.ValueExtractor.Field.ValueString()
					model.Labels[i].ValueExtractor.Field = &field
				}
				if !label.ValueExtractor.JSONPath.IsNull() {
					jsonPath := label.ValueExtractor.JSONPath.ValueString()
					model.Labels[i].ValueExtractor.JSONPath = &jsonPath
				}
				if !label.ValueExtractor.Regex.IsNull() {
					regex := label.ValueExtractor.Regex.ValueString()
					model.Labels[i].ValueExtractor.Regex = &regex
				}
			}
		}
	}

	// Convert filter
	if m.Filter != nil {
		model.Filter = &clientmodels.LogFilter{}
		if m.Filter.Match != nil {
			model.Filter.Match = m.toClientModelFilterMatch(m.Filter.Match)
		}
		if len(m.Filter.All) > 0 {
			model.Filter.MatchAll = &clientmodels.MatchAll{
				All: make([]*clientmodels.LogFilter, len(m.Filter.All)),
			}
			for i, allElem := range m.Filter.All {
				if allElem.Match != nil {
					model.Filter.MatchAll.All[i] = &clientmodels.LogFilter{
						Match: m.toClientModelFilterMatch(allElem.Match),
					}
				}
				if allElem.Not != nil && allElem.Not.Match != nil {
					model.Filter.MatchAll.All[i].Not = &clientmodels.LogFilter{
						Match: m.toClientModelFilterMatch(allElem.Not.Match),
					}
				}
			}
		}
		if len(m.Filter.Any) > 0 {
			model.Filter.MatchAny = &clientmodels.MatchAny{
				Any: make([]*clientmodels.LogFilter, len(m.Filter.Any)),
			}
			for i, filter := range m.Filter.Any {
				tflog.Debug(ctx, "filter", map[string]any{
					"filter": filter,
				})
				if filter.Match != nil {
					model.Filter.MatchAny.Any[i] = &clientmodels.LogFilter{
						Match: m.toClientModelFilterMatch(filter.Match),
					}
				}
				if filter.Not != nil && filter.Not.Match != nil {
					model.Filter.MatchAny.Any[i] = &clientmodels.LogFilter{
						MatchNone: &clientmodels.MatchNone{
							Not: &clientmodels.LogFilter{
								Match: m.toClientModelFilterMatch(filter.Not.Match),
							},
						},
					}
				}
				if filter.All != nil && len(filter.All) > 0 {
					model.Filter.MatchAny.Any[i] = &clientmodels.LogFilter{
						MatchAll: &clientmodels.MatchAll{
							All: make([]*clientmodels.LogFilter, len(filter.All)),
						},
					}
					for j, allElem := range filter.All {
						model.Filter.MatchAny.Any[i].MatchAll.All[j] = &clientmodels.LogFilter{
							Match: m.toClientModelFilterMatch(allElem.Match),
						}
					}
				}
			}
		}
		if m.Filter.Not != nil && m.Filter.Not.Match != nil {
			model.Filter.MatchNone = &clientmodels.MatchNone{
				Not: &clientmodels.LogFilter{
					Match: m.toClientModelFilterMatch(m.Filter.Not.Match),
				},
			}
		}
	}

	// Convert metric definitions
	if len(m.MetricDefinitions) > 0 {
		model.MetricDefinitions = make([]*clientmodels.MetricDefinition, len(m.MetricDefinitions))
		for i, def := range m.MetricDefinitions {
			model.MetricDefinitions[i] = &clientmodels.MetricDefinition{
				Name: def.Name.ValueString(),
				Type: clientmodels.MetricType(def.Type.ValueString()),
			}
			if !def.Field.IsNull() {
				model.MetricDefinitions[i].Field = def.Field.ValueString()
			}
			if !def.JSONPath.IsNull() {
				jsonPath := def.JSONPath.ValueString()
				model.MetricDefinitions[i].JSONPath = &jsonPath
			}
			if !def.Regex.IsNull() {
				regex := def.Regex.ValueString()
				model.MetricDefinitions[i].Regex = &regex
			}
		}
	}

	return nil
}

func (m *logMetricsResourceModel) toClientModelFilterMatch(match *matchModel) *clientmodels.Match {
	res := &clientmodels.Match{
		Field:    match.Field.ValueString(),
		Operator: clientmodels.MatchOperator(match.Operator.ValueString()),
	}

	if !match.JSONPath.IsNull() {
		jsonPath := match.JSONPath.ValueString()
		res.JSONPath = &jsonPath
	}

	if !match.Value.IsNull() {
		res.Value = match.Value.ValueString()
	}

	return res
}

func (m *logMetricsResourceModel) fromClientModelMatchModel(match *clientmodels.Match) *matchModel {
	res := &matchModel{
		Field:    types.StringValue(match.Field),
		Operator: types.StringValue(string(match.Operator)),
	}
	if match.JSONPath != nil {
		res.JSONPath = types.StringValue(*match.JSONPath)
	}
	if match.Value != "" {
		res.Value = types.StringValue(match.Value)
	}
	return res
}
