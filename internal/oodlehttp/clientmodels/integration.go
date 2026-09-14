package clientmodels

// IntegrationsResourcePath is the URL segment under
// /v1/api/instance/{instance}/ that the backend uses for integration CRUD.
// Every integration shares this one endpoint; which kind is being managed is
// carried in the body via the `type` discriminator and the matching
// `typeSpecificData` key, so each integration resource sends its own
// *IntegrationType constant rather than using a distinct path.
const IntegrationsResourcePath = "integrations"
