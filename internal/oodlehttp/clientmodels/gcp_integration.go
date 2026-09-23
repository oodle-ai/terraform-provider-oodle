package clientmodels

// GcpIntegrationType is the integration type discriminator value sent and
// received by the Oodle API for GCP metric-pull integrations.
const GcpIntegrationType = "GCP"

// DefaultOodlePrincipal is the Oodle service account that a customer grants
// the Service Account Token Creator role on their own service account. Oodle
// impersonates the customer's account through it, so no key ever leaves the
// customer project. The backend stores whatever value it is sent, so the
// default is applied here rather than server-side.
const DefaultOodlePrincipal = "gcp-monitoring-integration@oodle-ai.iam.gserviceaccount.com"

// GcpIntegration represents an Oodle GCP metric-pull integration over the
// wire. It is shaped to be both marshaled into
// CreateIntegrationRequest/PatchIntegration bodies and unmarshaled from the
// Integration response object returned by the backend; fields the request
// shape does not carry (CreatedAt, etc.) are simply ignored by the server.
type GcpIntegration struct {
	ID               string              `json:"id,omitempty"`
	Type             string              `json:"type,omitempty"`
	Name             string              `json:"name,omitempty"`
	Status           string              `json:"status,omitempty"`
	TypeSpecificData gcpTypeSpecificData `json:"typeSpecificData"`
}

// gcpTypeSpecificData mirrors the embedded GCPIntegration produced by the
// backend's OneOfIntegrationTypeSpecificData. Unlike the AWS and Azure
// wrappers there is no named key: projects sits directly under
// typeSpecificData. The response also carries empty sibling wrappers for the
// other integration types, which unmarshal away harmlessly.
type gcpTypeSpecificData struct {
	Projects []GcpProject `json:"projects,omitempty"`
}

// GcpProject mirrors the backend struct of the same name. Field order and
// JSON tags must match
// api-server/apps/integrations/models/gcp_integration.go.
//
// The array holds more than one project for rows written before a project
// became addressable on its own. A row written by this provider holds exactly
// one, which is what gives the project an id that Terraform can address.
type GcpProject struct {
	Status                 string `json:"status,omitempty"`
	Project                string `json:"project,omitempty"`
	CustomerServiceAccount string `json:"customerServiceAccount,omitempty"`
	OodlePrincipal         string `json:"oodlePrincipal,omitempty"`
}

// GetID returns the integration ID.
func (g *GcpIntegration) GetID() string {
	return g.ID
}
