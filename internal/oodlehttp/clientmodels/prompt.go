package clientmodels

// Prompt represents one version of a GenAI prompt.
//
// A version is immutable. Publishing a change means creating the
// next version and moving a label to it, so `Version` is
// assigned by the server and is only ever read back.
//
// The number is per deployment: it counts up within one
// instance's own store, so the same text is version 5 on one
// deployment and 4 on another as soon as one of them takes a
// change the other did not. `CommitMessage` is what identifies
// one definition across deployments.
type Prompt struct {
	Name          string   `json:"name"`
	Type          string   `json:"type,omitempty"`
	Prompt        string   `json:"prompt"`
	Labels        []string `json:"labels,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	CommitMessage string   `json:"commitMessage,omitempty"`
	Version       int      `json:"version,omitempty"`
}

// GetID returns the name, which is the fetch key callers use and
// the only stable identifier a prompt has. There is deliberately
// no rename on the API.
func (p *Prompt) GetID() string {
	return p.Name
}
