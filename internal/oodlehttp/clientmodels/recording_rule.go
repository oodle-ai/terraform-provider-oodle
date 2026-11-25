package clientmodels

// RecordingRule evaluates a PromQL expression on an interval and stores
// the result as a new metric.
type RecordingRule struct {
	// Slug is the unique ID of the rule in the instance. It cannot change.
	Slug string `json:"slug,omitempty"`

	// Name is the display name of the rule.
	Name string `json:"name"`

	// MetricName is the name of the metric that the rule writes.
	MetricName string `json:"metric_name"`

	// PrometheusExpr is the PromQL expression that the rule evaluates.
	PrometheusExpr string `json:"prometheus_expr"`

	// IntervalSecs is how often the rule runs. 0 uses the server default.
	IntervalSecs int64 `json:"interval_secs,omitempty"`

	// ExecutionGroup names the rules that run together, one after another.
	ExecutionGroup string `json:"execution_group,omitempty"`

	// LabelPolicy sets labels on every series that the rule writes.
	LabelPolicy *RecordingRuleLabelPolicy `json:"label_policy,omitempty"`
}

// RecordingRuleLabelPolicy sets labels on every series that a recording
// rule writes.
type RecordingRuleLabelPolicy struct {
	Add map[string]string `json:"add,omitempty"`
}

// RecordingRuleEnvelope is the request and response body of the
// recording rules API, which holds the rule under "recording_rule".
type RecordingRuleEnvelope struct {
	RecordingRule *RecordingRule `json:"recording_rule"`
}

// GetID returns the slug of the rule.
func (e *RecordingRuleEnvelope) GetID() string {
	if e.RecordingRule == nil {
		return ""
	}
	return e.RecordingRule.Slug
}
