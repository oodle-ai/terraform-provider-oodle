package clientmodels

import (
	amlabels "github.com/prometheus/alertmanager/pkg/labels"
)

type LabelMatcher struct {
	Type  amlabels.MatchType `json:"type"`
	Name  string             `json:"name"`
	Value string             `json:"value"`
}

// LabelMatcherNotificationPolicy defines a notification policy that is applied when
// alert labels match the specified matchers.
type LabelMatcherNotificationPolicy struct {
	// Matchers are the label matchers that determine when this policy applies
	Matchers []LabelMatcher `json:"matchers"`
	// NotificationPolicyID references the notification policy to use when labels match
	NotificationPolicyID ID `json:"notification_policy_id"`
}

// LabelMatcherNotifications defines a notification policy that is applied when
// alert labels match the specified matchers.
type LabelMatcherNotifications struct {
	// Matchers are the label matchers that determine when this policy applies
	Matchers []LabelMatcher `json:"matchers"`
	// NotificationPolicyID references the notification policy to use when labels match
	NotificationPolicyID ID `json:"notification_policy_id,omitempty"`
	// Notifiers are the notifiers for the policy.
	Notifiers NotifiersByCondition `json:"notifiers,omitempty"`
}
