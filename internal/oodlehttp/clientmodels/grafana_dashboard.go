package clientmodels

import jsoniter "github.com/json-iterator/go"

// GrafanaDashboard represents a Grafana dashboard for create/update operations.
type GrafanaDashboard struct {
	Dashboard interface{} `json:"dashboard"`
	FolderUID string      `json:"folderUid,omitempty"`
	Overwrite bool        `json:"overwrite,omitempty"`
	Message   string      `json:"message,omitempty"`
}

// GetID returns the dashboard UID from the dashboard JSON.
func (d *GrafanaDashboard) GetID() string {
	if dashMap, ok := d.Dashboard.(map[string]interface{}); ok {
		if uid, ok := dashMap["uid"].(string); ok {
			return uid
		}
	}
	return ""
}

// GrafanaDashboardResponse represents the response from saving a dashboard.
type GrafanaDashboardResponse struct {
	ID      int    `json:"id"`
	UID     string `json:"uid"`
	URL     string `json:"url"`
	Status  string `json:"status"`
	Version int    `json:"version"`
	Slug    string `json:"slug"`
}

// GetID returns the UID as the identifier for the dashboard.
func (r *GrafanaDashboardResponse) GetID() string {
	return r.UID
}

// GrafanaDashboardGetResponse represents the response from getting a dashboard.
type GrafanaDashboardGetResponse struct {
	Dashboard interface{}   `json:"dashboard"`
	Meta      DashboardMeta `json:"meta"`
}

// GetID returns the dashboard UID from the dashboard JSON.
func (r *GrafanaDashboardGetResponse) GetID() string {
	if dashMap, ok := r.Dashboard.(map[string]interface{}); ok {
		if uid, ok := dashMap["uid"].(string); ok {
			return uid
		}
	}
	return ""
}

// GetConfigJSON returns the dashboard JSON as a string.
func (r *GrafanaDashboardGetResponse) GetConfigJSON() (string, error) {
	bytes, err := jsoniter.Marshal(r.Dashboard)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// DashboardMeta represents metadata for a dashboard.
type DashboardMeta struct {
	FolderID    int    `json:"folderId"`
	FolderUID   string `json:"folderUid"`
	FolderTitle string `json:"folderTitle"`
	FolderURL   string `json:"folderUrl"`
	URL         string `json:"url"`
	Version     int    `json:"version"`
	Slug        string `json:"slug"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}
