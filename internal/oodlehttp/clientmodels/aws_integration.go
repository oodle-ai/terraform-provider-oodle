package clientmodels

// AwsIntegrationType is the integration type discriminator value sent and
// received by the Oodle API for AWS CloudWatch metric-pull integrations.
const AwsIntegrationType = "CLOUDWATCH_METRIC_PULL"

// AwsIntegration represents an Oodle AWS CloudWatch metric-pull integration
// over the wire. It is shaped to be both marshaled into
// CreateIntegrationRequest/PatchIntegration bodies and unmarshaled from the
// Integration response object returned by the backend; fields the request
// shape does not carry (CreatedAt, etc.) are simply ignored by the server.
type AwsIntegration struct {
	ID               string                 `json:"id,omitempty"`
	Type             string                 `json:"type,omitempty"`
	Name             string                 `json:"name,omitempty"`
	Status           string                 `json:"status,omitempty"`
	TypeSpecificData awsTypeSpecificData    `json:"typeSpecificData"`
}

// awsTypeSpecificData mirrors the embedded
// CloudWatchMetricPullIntegrationWrapper produced by the backend's
// OneOfIntegrationTypeSpecificData: the actual config lives under the
// cloudWatchMetricPullIntegration key.
type awsTypeSpecificData struct {
	CloudWatchMetricPullIntegration CloudWatchMetricPullIntegration `json:"cloudWatchMetricPullIntegration"`
}

// CloudWatchMetricPullIntegration mirrors the backend struct of the same
// name. Field order and JSON tags must match
// api-server/apps/integrations/models/cloudwatch_metric_pull_integration.go.
type CloudWatchMetricPullIntegration struct {
	AccountID                   string                             `json:"accountId,omitempty"`
	LaunchCFStackURL            string                             `json:"launchCFStackURL,omitempty"`
	LaunchCFStackRegion         string                             `json:"launchCFStackRegion,omitempty"`
	RoleArn                     string                             `json:"roleArn,omitempty"`
	ExternalID                  string                             `json:"externalId,omitempty"`
	Regions                     []string                           `json:"regions,omitempty"`
	ResourceTypesSearchTagsList []CloudWatchResourceTypeSearchTags `json:"resourceTypesSearchTagsList,omitempty"`
}

// CloudWatchResourceTypeSearchTags pairs a set of CloudWatch resource types
// (e.g. AWS/EC2) with optional search-tag filters applied during discovery.
type CloudWatchResourceTypeSearchTags struct {
	ResourceTypes []string              `json:"resourceTypes,omitempty"`
	SearchTags    []CloudWatchSearchTag `json:"searchTags,omitempty"`
}

// CloudWatchSearchTag is a single tag filter. Value may be a regex.
type CloudWatchSearchTag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// GetID returns the integration ID.
func (a *AwsIntegration) GetID() string {
	return a.ID
}
