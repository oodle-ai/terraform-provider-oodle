package clientmodels

// GrafanaFolder represents a Grafana folder in the API.
type GrafanaFolder struct {
	ID        int    `json:"id,omitempty"`
	UID       string `json:"uid,omitempty"`
	Title     string `json:"title"`
	URL       string `json:"url,omitempty"`
	ParentUID string `json:"parentUid,omitempty"`
	Version   int    `json:"version,omitempty"`
	Created   string `json:"created,omitempty"`
	Updated   string `json:"updated,omitempty"`
}

// GetID returns the UID as the identifier for the folder.
func (f *GrafanaFolder) GetID() string {
	return f.UID
}
