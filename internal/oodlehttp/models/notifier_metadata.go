package models

import (
	"github.com/cockroachdb/errors"
)

type NotifierType int16

const (
	NotifierConfigEmail NotifierType = iota
	NotifierConfigPagerduty
	NotifierConfigSlack
	NotifierConfigOpsGenie
	NotifierConfigWebhook
)

var notifierTypeToName = map[NotifierType]string{
	NotifierConfigEmail:     "email",
	NotifierConfigPagerduty: "pagerduty",
	NotifierConfigSlack:     "slack",
	NotifierConfigOpsGenie:  "opsgenie",
	NotifierConfigWebhook:   "webhook",
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
