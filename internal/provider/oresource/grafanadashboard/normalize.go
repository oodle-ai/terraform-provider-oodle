package grafanadashboard

import (
	"encoding/json"
	"reflect"
)

// serverManagedKeys are dashboard JSON fields Grafana sets on every save.
// They never come from the user's config, so leaving them in config_json
// makes every plan show a diff, and every apply bumps version again.
var serverManagedKeys = []string{"id", "version"}

// normalizeDashboardJSON removes the server-managed fields from a
// dashboard JSON document and returns the decoded result.
func normalizeDashboardJSON(raw string) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	for _, k := range serverManagedKeys {
		delete(m, k)
	}
	return m, nil
}

// configJSONForState picks the config_json value to store after a Read.
//
// remote is the dashboard JSON the API returned, prior is the config_json
// already in state (empty on import), and uid is the dashboard's UID.
// If remote matches prior once server-managed fields are dropped, prior is
// kept as-is so the user's formatting and key order don't cause a diff.
// Otherwise the normalized remote JSON is returned, which shows real drift.
func configJSONForState(remote, prior, uid string) (string, error) {
	remoteMap, err := normalizeDashboardJSON(remote)
	if err != nil {
		return "", err
	}

	if prior != "" {
		if priorMap, err := normalizeDashboardJSON(prior); err == nil {
			// A config without a uid gets one assigned on create.
			// That's not drift, so ignore it when comparing.
			cmp := remoteMap
			if _, ok := priorMap["uid"]; !ok && remoteMap["uid"] == uid {
				cmp = make(map[string]interface{}, len(remoteMap))
				for k, v := range remoteMap {
					if k != "uid" {
						cmp[k] = v
					}
				}
			}
			if reflect.DeepEqual(cmp, priorMap) {
				return prior, nil
			}
		}
	}

	out, err := json.Marshal(remoteMap)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
