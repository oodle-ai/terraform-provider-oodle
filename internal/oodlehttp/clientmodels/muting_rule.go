package clientmodels

// MutingRule suppresses notifications for the alerts its matchers
// select.
//
// A rule is either one-off, bounded by StartsAt and EndsAt, or
// recurring, driven by ScheduleIDs. The two are mutually exclusive.
// Every rule must carry an equality matcher on _oodle_monitor_id:
// muting is scoped to a monitor, and the UI relies on that to attach
// rules to the monitor they mute.
type MutingRule struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Comment string `json:"comment,omitempty"`
	// Matchers select the alerts to mute. An alert is muted when it
	// matches all of them.
	Matchers []LabelMatcher `json:"matchers"`
	// ScheduleIDs makes the rule recurring: the alert is muted during
	// any of the referenced schedules.
	ScheduleIDs []string `json:"scheduleIds,omitempty"`
	// StartsAt and EndsAt are RFC 3339 timestamps bounding a one-off
	// rule. An unset EndsAt mutes forever.
	StartsAt  string `json:"startsAt,omitempty"`
	EndsAt    string `json:"endsAt,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

func (m *MutingRule) GetID() string { return m.ID }
