package clientmodels

// MetricDropRule represents a rule for dropping metric time-series at ingest time.
type MetricDropRule struct {
	// ID is the unique identifier.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// RuleName is the human-readable name for the drop rule.
	RuleName string `json:"rule_name,omitempty" yaml:"rule_name,omitempty"`

	// Type is the type of the drop rule.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// MetricName is the __name__ label matcher that selects which metrics to drop.
	MetricName *LabelMatcher `json:"metric_name,omitempty" yaml:"metric_name,omitempty"`

	// Filters are optional additional label matchers that further restrict which series are dropped.
	Filters []*LabelMatcher `json:"filters" yaml:"filters"`
}

// GetID returns the ID of the metric drop rule.
func (r *MetricDropRule) GetID() string {
	return r.ID
}
