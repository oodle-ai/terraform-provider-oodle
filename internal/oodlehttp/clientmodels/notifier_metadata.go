package clientmodels

import (
	"github.com/cockroachdb/errors"
)

type NotifierType int16

// Values must match the notifier type enum of the Oodle API and are therefore
// spelled out explicitly rather than derived from iota.
const (
	NotifierConfigEmail      NotifierType = 0
	NotifierConfigPagerduty  NotifierType = 1
	NotifierConfigSlack      NotifierType = 2
	NotifierConfigOpsGenie   NotifierType = 3
	NotifierConfigWebhook    NotifierType = 4
	NotifierConfigGoogleChat NotifierType = 5
	NotifierConfigMSTeamsV2  NotifierType = 6
	NotifierConfigRootly     NotifierType = 7
)

var notifierTypeToName = map[NotifierType]string{
	NotifierConfigEmail:      "email",
	NotifierConfigPagerduty:  "pagerduty",
	NotifierConfigSlack:      "slack",
	NotifierConfigOpsGenie:   "opsgenie",
	NotifierConfigWebhook:    "webhook",
	NotifierConfigGoogleChat: "googlechat",
	NotifierConfigMSTeamsV2:  "msteamsv2",
	NotifierConfigRootly:     "rootly",
}

var notifierNameToType = map[string]NotifierType{}

var NotifierNames map[string]struct{}

func (nt NotifierType) AsInt16() int16 {
	return int16(nt)
}

func GetNotifierType(name string) (NotifierType, error) {
	if t, ok := notifierNameToType[name]; !ok {
		return 0, errors.Newf("invalid notifier type: %s", name)
	} else {
		return t, nil
	}
}

func (nt NotifierType) String() string {
	return notifierTypeToName[nt]
}

func NewNotifierTypeFromString(val string) (NotifierType, error) {
	for k, v := range notifierTypeToName {
		if v == val {
			return k, nil
		}
	}

	return 0, errors.Newf("invalid notifier type: %s", val)
}

func init() {
	NotifierNames = make(map[string]struct{})
	for t, v := range notifierTypeToName {
		NotifierNames[v] = struct{}{}
		notifierNameToType[v] = t
	}
}
