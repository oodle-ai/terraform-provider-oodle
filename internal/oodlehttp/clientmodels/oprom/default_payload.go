package oprom

import "strings"

// DefaultPayloadField is one key of the body Alertmanager posts when a notifier
// sets no custom payload, with the template that reproduces it.
type DefaultPayloadField struct {
	Key      string
	Template string
}

// DefaultPayloadFields lists the default body, in the order Alertmanager
// documents it. A custom payload replaces the body outright, thus a notifier
// whose receiver parses the standard shape (Rootly) must carry these keys
// itself once it sets one.
//
// Only the notification data is in scope while the payload renders, thus
// `version` and `truncatedAlerts` are the constants they always are (Oodle
// pins max_alerts to 0). Two fields of the real body cannot be reproduced:
// `groupKey` is out of scope for every payload template, and `version` arrives
// as the number 4 rather than the string "4", because a rendered string that
// parses as YAML is inlined as structured data.
var DefaultPayloadFields = []DefaultPayloadField{
	{Key: "receiver", Template: "{{ .Receiver }}"},
	{Key: "status", Template: "{{ .Status }}"},
	{Key: "alerts", Template: "{{ .Alerts | toJson }}"},
	{Key: "groupLabels", Template: "{{ .GroupLabels | toJson }}"},
	{Key: "commonLabels", Template: "{{ .CommonLabels | toJson }}"},
	{Key: "commonAnnotations", Template: "{{ .CommonAnnotations | toJson }}"},
	{Key: "externalURL", Template: "{{ .ExternalURL }}"},
	{Key: "version", Template: "4"},
	{Key: "truncatedAlerts", Template: "0"},
}

// DefaultPayloadKeys lists the default keys for a schema description, so the
// documentation cannot drift from the list above.
func DefaultPayloadKeys() string {
	keys := make([]string, 0, len(DefaultPayloadFields))
	for _, field := range DefaultPayloadFields {
		keys = append(keys, "`"+field.Key+"`")
	}
	return strings.Join(keys, ", ")
}
