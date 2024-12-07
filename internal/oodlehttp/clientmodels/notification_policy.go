package clientmodels

// NotificationPolicy represents a policy for sending notifications based on severity.
// A notification policy is associated with a monitor.
type NotificationPolicy struct {
	ID        ID                  `json:"id,omitempty" yaml:"id,omitempty"`
	Name      string              `json:"name,omitempty" yaml:"name,omitempty"`
	Notifiers NotifiersBySeverity `json:"notifiers,omitempty" yaml:"notifiers,omitempty"`
	// Global policy is applied to all monitors in addition to any monitor specific policies.
	Global     bool `json:"global,omitempty" yaml:"global,omitempty"`
	MuteGlobal bool `json:"mute_global,omitempty" yaml:"mute_global,omitempty"`
	// MuteNonGlobal is used to disable all non-global policies. It can only be set for a Global
	// notification policy. Global policy would still be effective when MuteNonGlobal is true.
	MuteNonGlobal bool `json:"mute_non_global,omitempty" yaml:"mute_non_global,omitempty"`
}

// NotifiersBySeverity represents notifiers for each severity level.
type NotifiersBySeverity struct {
	Warn     []ID `json:"warn,omitempty" yaml:"warn,omitempty"`
	Critical []ID `json:"critical,omitempty" yaml:"critical,omitempty"`
}
