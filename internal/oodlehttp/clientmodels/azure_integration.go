package clientmodels

// AzureIntegrationType is the integration type discriminator value sent and
// received by the Oodle API for Azure Monitor metric-pull integrations.
const AzureIntegrationType = "AZURE_METRICS"

// AzureSecretMask is what the backend returns in place of the real client
// secret on every read. Sending it back unchanged on update tells the backend
// to keep the stored secret; see
// api-server/apps/integrations/models/azure_metrics_integration.go.
const AzureSecretMask = "********"

// AzureIntegration represents an Oodle Azure Monitor metric-pull integration
// over the wire. It is shaped to be both marshaled into
// CreateIntegrationRequest/PatchIntegration bodies and unmarshaled from the
// Integration response object returned by the backend; fields the request
// shape does not carry (CreatedAt, etc.) are simply ignored by the server.
type AzureIntegration struct {
	ID               string                `json:"id,omitempty"`
	Type             string                `json:"type,omitempty"`
	Name             string                `json:"name,omitempty"`
	Status           string                `json:"status,omitempty"`
	TypeSpecificData azureTypeSpecificData `json:"typeSpecificData"`
}

// azureTypeSpecificData mirrors the embedded AzureMetricsIntegrationWrapper
// produced by the backend's OneOfIntegrationTypeSpecificData: the actual
// config lives under the azureMetricsIntegration key. The response also
// carries empty sibling wrappers for the other integration types, which
// unmarshal away harmlessly.
type azureTypeSpecificData struct {
	AzureMetricsIntegration AzureMetricsIntegration `json:"azureMetricsIntegration"`
}

// AzureMetricsIntegration mirrors the backend struct of the same name. Field
// order and JSON tags must match
// api-server/apps/integrations/models/azure_metrics_integration.go.
//
// One record is one Azure subscription. ClientSecret is write-only: it is
// accepted on create and update, but reads return AzureSecretMask instead of
// the stored value.
type AzureMetricsIntegration struct {
	SubscriptionName  string               `json:"subscriptionName,omitempty"`
	TenantID          string               `json:"tenantId,omitempty"`
	SubscriptionID    string               `json:"subscriptionId,omitempty"`
	ClientID          string               `json:"clientId,omitempty"`
	ClientSecret      string               `json:"clientSecret,omitempty"`
	ServiceFilters    []AzureServiceFilter `json:"serviceFilters,omitempty"`
	ResourceGroups    []string             `json:"resourceGroups,omitempty"`
	ResourceNameRegex string               `json:"resourceNameRegex,omitempty"`
}

// AzureServiceFilter is one group of Azure services plus the tags that narrow
// it. Tags are include-only: a resource must carry every tag listed to be
// collected by this group.
type AzureServiceFilter struct {
	ServiceIDs []string          `json:"serviceIds,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// GetID returns the integration ID.
func (a *AzureIntegration) GetID() string {
	return a.ID
}
