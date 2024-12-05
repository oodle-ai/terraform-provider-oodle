package models

// Grouping is a model for grouping alerts.
type Grouping struct {
	ByMonitor bool     `json:"by_monitor,omitempty" yaml:"by_monitor,omitempty"`
	ByLabels  []string `json:"by_labels,omitempty" yaml:"by_labels,omitempty"`
	Disabled  bool     `json:"disabled,omitempty" yaml:"disabled,omitempty"`
}
