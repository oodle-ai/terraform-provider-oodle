package clientmodels

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/common/model"
)

// Monitor is a model for a monitor.
type Monitor struct {
	ID ID `json:"id,omitempty" yaml:"id,omitempty"`
	// Name is the name of the monitor.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Interval is the interval at which the monitor should be evaluated.
	Interval time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`
	// PromQLQuery is the Prometheus query for the monitor.
	PromQLQuery string `json:"promql_query,omitempty" yaml:"promql_query,omitempty"`
	// Conditions are the conditions for the monitor for each severity level.
	Conditions ConditionBySeverity `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	// Labels are the labels for the monitor.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// Annotations are the annotations for the monitor.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// Grouping is the grouping configuration for the monitor.
	Grouping Grouping `json:"grouping,omitempty" yaml:"grouping,omitempty"`
	// Deprecated: Use Notifications instead.
	// NotificationPolicyID is the ID of the notification policy associated with the monitor.
	// It is an optional field.
	NotificationPolicyID *ID `json:"notification_policy_id,omitempty" yaml:"notification_policy_id,omitempty"`
	// Deprecated: Use Notifications instead.
	// LabelMatcherNotificationPolicies is the list of label matcher notification policies for the monitor.
	// These policies are evaluated in order, and the first matching policy is used. Within a label matcher,
	// all matchers must match for policy to be effective.
	// If no policy matches, the default NotificationPolicyID is used if set.
	LabelMatcherNotificationPolicies []LabelMatcherNotificationPolicy `json:"label_matcher_notification_policies,omitempty" yaml:"label_matcher_notification_policies,omitempty"`
	// Notifications is the list of notifications for the monitor.
	// These notifications are evaluated in order, and the first matching notification is used.
	Notifications []LabelMatcherNotifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	// GroupWait is the time to wait before sending the first alert for a group of alerts.
	GroupWait *time.Duration `json:"group_wait,omitempty" yaml:"group_wait,omitempty"`
	// GroupInterval is the interval at which to send alerts for the same group of alerts after the first alert.
	GroupInterval *time.Duration `json:"group_interval,omitempty" yaml:"group_interval,omitempty"`
	// RepeatInterval is the interval at which to send alerts for the same alert after firing.
	// RepeatInterval should be a multiple of GroupInterval
	RepeatInterval *time.Duration `json:"repeat_interval,omitempty" yaml:"repeat_interval,omitempty"`
}

var _ ClientModel = (*Monitor)(nil)

// MarshalJSON customizes the JSON marshaling for Monitor.
func (m Monitor) MarshalJSON() ([]byte, error) {
	type Alias Monitor
	return jsoniter.Marshal(&struct {
		*Alias
		Interval       model.Duration  `json:"interval,omitempty"`
		GroupWait      *model.Duration `json:"group_wait,omitempty"`
		GroupInterval  *model.Duration `json:"group_interval,omitempty"`
		RepeatInterval *model.Duration `json:"repeat_interval,omitempty"`
	}{
		Alias:          (*Alias)(&m),
		Interval:       *toPromDuration(&m.Interval),
		GroupWait:      toPromDuration(m.GroupWait),
		GroupInterval:  toPromDuration(m.GroupInterval),
		RepeatInterval: toPromDuration(m.RepeatInterval),
	})
}

// UnmarshalJSON customizes the JSON unmarshaling for Monitor.
func (m *Monitor) UnmarshalJSON(data []byte) error {
	type Alias Monitor
	aux := &struct {
		*Alias
		Interval       model.Duration  `json:"interval,omitempty"`
		GroupWait      *model.Duration `json:"group_wait,omitempty"`
		GroupInterval  *model.Duration `json:"group_interval,omitempty"`
		RepeatInterval *model.Duration `json:"repeat_interval,omitempty"`
	}{
		Alias: (*Alias)(m),
	}
	if err := jsoniter.Unmarshal(data, aux); err != nil {
		return err
	}

	m.Interval = time.Duration(aux.Interval)
	m.GroupWait = FromPromDuration(aux.GroupWait)
	m.GroupInterval = FromPromDuration(aux.GroupInterval)
	m.RepeatInterval = FromPromDuration(aux.RepeatInterval)
	return nil
}

func (m Monitor) GetID() string {
	return m.ID.UUID.String()
}

func toPromDuration(d *time.Duration) *model.Duration {
	if d == nil {
		return nil
	}

	pd := model.Duration(*d)
	return &pd
}

// FromPromDuration converts a Prometheus duration to a time.Duration.
func FromPromDuration(d *model.Duration) *time.Duration {
	if d == nil {
		return nil
	}
	dur := time.Duration(*d)
	return &dur
}
