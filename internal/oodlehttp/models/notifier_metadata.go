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

var notifierTypeToname = map[NotifierType]string{
	NotifierConfigEmail:     "email",
	NotifierConfigPagerduty: "pagerduty",
	NotifierConfigSlack:     "slack",
	NotifierConfigOpsGenie:  "opsgenie",
	NotifierConfigWebhook:   "webhook",
}

var NotifierNames map[string]struct{}

func (nt NotifierType) AsInt16() int16 {
	return int16(nt)
}

func NotifierTypeFromInt16(val int16) (NotifierType, error) {
	switch val {
	case NotifierConfigEmail.AsInt16():
		return NotifierConfigEmail, nil
	case NotifierConfigPagerduty.AsInt16():
		return NotifierConfigPagerduty, nil
	case NotifierConfigSlack.AsInt16():
		return NotifierConfigSlack, nil
	case NotifierConfigOpsGenie.AsInt16():
		return NotifierConfigOpsGenie, nil
	case NotifierConfigWebhook.AsInt16():
		return NotifierConfigWebhook, nil
	default:
		return 0, errors.Newf("invalid notifier type: %d", val)
	}
}

func (nt NotifierType) String() string {
	return notifierTypeToname[nt]
}

func NewNotifierTypeFromString(val string) (NotifierType, error) {
	for k, v := range notifierTypeToname {
		if v == val {
			return k, nil
		}
	}

	return 0, errors.Newf("invalid notifier type: %s", val)
}

func init() {
	NotifierNames = make(map[string]struct{})
	for _, v := range notifierTypeToname {
		NotifierNames[v] = struct{}{}
	}
}
