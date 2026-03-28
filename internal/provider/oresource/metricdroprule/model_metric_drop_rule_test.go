package metricdroprule

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	amlabels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestMetricDropRuleModel(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.MetricDropRule{
		ID:       "test-id-123",
		RuleName: "Drop unused go_gc metrics",
		Type:     "drop",
		MetricName: &clientmodels.DropRuleMatcher{
			Name:  "__name__",
			Type:  amlabels.MatchRegexp,
			Value: "go_gc_.*",
		},
		Filters: []*clientmodels.DropRuleMatcher{
			{
				Name:  "job",
				Type:  amlabels.MatchEqual,
				Value: "unused-exporter",
			},
			{
				Name:  "cluster",
				Type:  amlabels.MatchNotEqual,
				Value: "production",
			},
		},
	}

	resourceModel := &metricDropRuleResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.MetricDropRule{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}

func TestMetricDropRuleModelNoFilters(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.MetricDropRule{
		ID:       "test-id-456",
		RuleName: "Drop all kube_state metrics",
		Type:     "drop",
		MetricName: &clientmodels.DropRuleMatcher{
			Name:  "__name__",
			Type:  amlabels.MatchEqual,
			Value: "kube_state_metrics_total",
		},
	}

	resourceModel := &metricDropRuleResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.MetricDropRule{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}

func TestMetricDropRuleModelAllMatchTypes(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.MetricDropRule{
		ID:       "test-id-789",
		RuleName: "Test all match types",
		Type:     "drop",
		MetricName: &clientmodels.DropRuleMatcher{
			Name:  "__name__",
			Type:  amlabels.MatchNotRegexp,
			Value: "important_.*",
		},
		Filters: []*clientmodels.DropRuleMatcher{
			{
				Name:  "env",
				Type:  amlabels.MatchEqual,
				Value: "staging",
			},
			{
				Name:  "team",
				Type:  amlabels.MatchNotEqual,
				Value: "platform",
			},
			{
				Name:  "service",
				Type:  amlabels.MatchRegexp,
				Value: "api-.*",
			},
			{
				Name:  "region",
				Type:  amlabels.MatchNotRegexp,
				Value: "us-.*",
			},
		},
	}

	resourceModel := &metricDropRuleResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.MetricDropRule{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}
