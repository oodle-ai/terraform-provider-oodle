package clientmodels

// LogMetrics is a definition to convert logs to metrics.
type LogMetrics struct {
	// ID is the unique identifier.
	ID ID `json:"id,omitempty" yaml:"id,omitempty"`

	// Name is the name of the rule.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Labels are the labels that will be added to all metrics created
	// by this configuration.
	Labels []*Label `json:"labels,omitempty" yaml:"labels,omitempty"`

	// Filter is the log filter used to determine which logs to process
	// to create the metrics.
	Filter *LogFilter `json:"filter,omitempty" yaml:"filter,omitempty"`

	// MetricDefinitions defines all the metrics to be created from the logs.
	MetricDefinitions []*MetricDefinition `json:"metricDefinitions,omitempty" yaml:"metricDefinitions,omitempty"`

	// UpdatedAtEpochMs is the updated at time in milliseconds since epoch.
	UpdatedAtEpochMs int64 `json:"updatedAtEpochMs,omitempty" yaml:"updatedAtEpochMs,omitempty"`
}

// Label represents a name-value pair for a metric label.
type Label struct {
	// Name is the name of the label.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Value is the static value of the label.
	// Only one of Value or ValueExtractor should be set.
	Value *string `json:"value,omitempty" yaml:"value,omitempty"`

	// ValueExtractor defines how to extract the label value.
	// Only one of Value or ValueExtractor should be set.
	ValueExtractor *ValueExtractor `json:"valueExtractor,omitempty" yaml:"valueExtractor,omitempty"`
}

// ValueExtractor is used to extract label values from the log fields.
type ValueExtractor struct {
	// Field is the name of the field in the log to extract the value from.
	Field *string `json:"field,omitempty" yaml:"field,omitempty"`

	// JSONPath specifies a path to extract a nested value from a JSON field.
	// If not set, the entire field value is used as the label value.
	JSONPath *string `json:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`

	// Regex is used to derive the label value by matching the
	// regex pattern. If not set, the entire field value is used as the
	// label value.
	Regex *string `json:"regex,omitempty" yaml:"regex,omitempty"`
}

// MatchOperator is the operator used to match in a match filter.
type MatchOperator string

const (
	IsOperator           MatchOperator = "is"
	ContainsOperator     MatchOperator = "contains"
	MatchesRegexOperator MatchOperator = "matches regex"
	ExistsOperator       MatchOperator = "exists"
)

// Match is a filter that matches a log field against a value.
type Match struct {
	// Field is the name of the log field to match against.
	Field string `json:"field,omitempty" yaml:"field,omitempty"`
	// JSONPath is used to match against a value at a specific path in the
	// JSON field. If not specified, matches against the entire field value.
	JSONPath *string `json:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`

	// Operator is the operator used to match the field value.
	Operator MatchOperator `json:"operator,omitempty" yaml:"operator,omitempty"`

	// Value is the value to match against.
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// MatchAll is a list of filters where all the filters must match.
type MatchAll struct {
	All []*LogFilter `json:"all,omitempty" yaml:"all,omitempty"`
}

// MatchAny is a list of filters where at least one filter must match.
type MatchAny struct {
	Any []*LogFilter `json:"any,omitempty" yaml:"any,omitempty"`
}

// MatchNone is a filter where none of the filters must match.
type MatchNone struct {
	Not *LogFilter `json:"not,omitempty" yaml:"not,omitempty"`
}

// LogFilter is for filtering logs using match conditions.
//
// It is an oneof type that can be one of the following:
// - Match
// - MatchAll
// - MatchAny
// - MatchNone.
type LogFilter struct {
	*Match
	*MatchAll
	*MatchAny
	*MatchNone
}

// MetricType is the type of the metric to be created.
type MetricType string

const (
	LogCountMetricDefinition  MetricType = "count"
	CounterMetricDefinition   MetricType = "counter"
	GaugeMetricDefinition     MetricType = "gauge"
	HistogramMetricDefinition MetricType = "histogram"
)

// MetricDefinition represents the definition of a metric to be created.
type MetricDefinition struct {
	// Name is the name of the metric to be created. Must match Prometheus metric naming rules.
	// See https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Type is the type of the metric to be created.
	Type MetricType `json:"type,omitempty" yaml:"type,omitempty"`

	// Field is the name of the log field to extract from. Only used when Type is
	// "counter" or "gauge".
	Field string `json:"field,omitempty" yaml:"field,omitempty"`

	// JSONPath is an optional JSON path to extract a numeric value from a JSON field.
	// Only used when Type is "counter" or "gauge". Cannot be used together with Regex.
	JSONPath *string `json:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`

	// Regex is an optional regex pattern to extract a numeric value from the field.
	// Only used when Type is "counter" or "gauge". Cannot be used together with JSONPath.
	Regex *string `json:"regex,omitempty" yaml:"regex,omitempty"`
}

// GetID returns the ID of the log metrics rule.
func (l *LogMetrics) GetID() string {
	return l.ID.UUID.String()
}
