package logmetrics

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestLogMetricsModel(t *testing.T) {
	ctx := context.Background()
	field := "field1"
	jsonPath := "$.path"
	regex := ".*"
	value := "value1"

	clientModel := &clientmodels.LogMetrics{
		ID: clientmodels.ID{
			UUID: uuid.New(),
		},
		Name: "test_metrics",
		Labels: []*clientmodels.Label{
			{
				Name:  "label1",
				Value: &value,
			},
			{
				Name: "label2",
				ValueExtractor: &clientmodels.ValueExtractor{
					Field:    &field,
					JSONPath: &jsonPath,
					Regex:    &regex,
				},
			},
		},
		Filter: &clientmodels.LogFilter{
			Match: &clientmodels.Match{
				Field:    "field1",
				JSONPath: &jsonPath,
				Operator: clientmodels.IsOperator,
				Value:    "value1",
			},
			MatchAll: &clientmodels.MatchAll{
				All: []*clientmodels.LogFilter{
					{
						Match: &clientmodels.Match{
							Field:    "field2",
							Operator: clientmodels.ContainsOperator,
							Value:    "value2",
						},
					},
				},
			},
			MatchAny: &clientmodels.MatchAny{
				Any: []*clientmodels.LogFilter{
					{
						Match: &clientmodels.Match{
							Field:    "field3",
							Operator: clientmodels.MatchesRegexOperator,
							Value:    ".*",
						},
					},
				},
			},
			MatchNone: &clientmodels.MatchNone{
				Not: &clientmodels.LogFilter{
					Match: &clientmodels.Match{
						Field:    "field4",
						Operator: clientmodels.ExistsOperator,
					},
				},
			},
		},
		MetricDefinitions: []*clientmodels.MetricDefinition{
			{
				Name:  "metric1",
				Type:  clientmodels.LogCountMetricDefinition,
				Field: "field1",
			},
			{
				Name:     "metric2",
				Type:     clientmodels.CounterMetricDefinition,
				Field:    "field2",
				JSONPath: &jsonPath,
			},
			{
				Name:  "metric3",
				Type:  clientmodels.GaugeMetricDefinition,
				Field: "field3",
				Regex: &regex,
			},
			{
				Name: "metric4",
				Type: clientmodels.HistogramMetricDefinition,
			},
		},
	}

	resourceModel := &logMetricsResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.LogMetrics{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}
