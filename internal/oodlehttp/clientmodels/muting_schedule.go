package clientmodels

// MutingSchedule is a recurring window that muting rules reference by
// id to mute on a repeating calendar rather than over a fixed span.
//
// An alert is muted while the current time falls inside any of the
// schedule's time intervals.
type MutingSchedule struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// TimeIntervals are the recurring windows. At least one is
	// required, and the schedule is active during any of them.
	TimeIntervals []ScheduleTimeInterval `json:"time_intervals"`
	CreatedBy     string                 `json:"created_by,omitempty"`
	CreatedAt     string                 `json:"created_at,omitempty"`
	UpdatedAt     string                 `json:"updated_at,omitempty"`
}

func (m *MutingSchedule) GetID() string { return m.ID }

// ScheduleTimeInterval is one recurring window, expressed as
// Alertmanager expresses them: a set of independent constraints that
// a moment must satisfy all of to fall inside the interval. An
// omitted constraint matches everything, so an interval carrying only
// Weekdays covers those days in full.
type ScheduleTimeInterval struct {
	Times       []ScheduleTimeRange `json:"times,omitempty"`
	Weekdays    []string            `json:"weekdays,omitempty"`
	DaysOfMonth []string            `json:"days_of_month,omitempty"`
	Months      []string            `json:"months,omitempty"`
	Years       []string            `json:"years,omitempty"`
	// Location is the IANA timezone the other constraints are read
	// in. The server defaults it to UTC.
	Location string `json:"location,omitempty"`
}

// ScheduleTimeRange is a span within a day, in 24-hour HH:MM.
type ScheduleTimeRange struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
