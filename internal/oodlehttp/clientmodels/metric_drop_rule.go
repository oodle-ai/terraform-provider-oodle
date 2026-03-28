package clientmodels

import (
	amlabels "github.com/prometheus/alertmanager/pkg/labels"
)

// MetricDropRule represents a rule for dropping metric time-series at ingest time.
type MetricDropRule struct {
	// ID is the unique identifier.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// RuleName is the human-readable name for the drop rule.
	RuleName string `json:"rule_name,omitempty" yaml:"rule_name,omitempty"`

	// Type is the type of the drop rule.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// MetricName is the __name__ label matcher that selects which metrics to drop.
	MetricName *DropRuleMatcher `json:"metric_name,omitempty" yaml:"metric_name,omitempty"`

	// Filters are optional additional label matchers that further restrict which series are dropped.
	Filters []*DropRuleMatcher `json:"filters" yaml:"filters"`
}

// DropRuleMatcher represents a label matcher used in drop rules.
type DropRuleMatcher struct {
	// Name is the label name to match against.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Type is the match type (=, !=, =~, !~).
	Type amlabels.MatchType `json:"type" yaml:"type"`

	// Value is the value to match against.
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// GetID returns the ID of the metric drop rule.
func (r *MetricDropRule) GetID() string {
	return r.ID
}
