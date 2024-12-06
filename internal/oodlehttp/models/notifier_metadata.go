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
