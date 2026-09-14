package azureintegration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type azureIntegrationResourceModel struct {
	ID                types.String              `tfsdk:"id"`
	Name              types.String              `tfsdk:"name"`
	Status            types.String              `tfsdk:"status"`
	SubscriptionName  types.String              `tfsdk:"subscription_name"`
	TenantID          types.String              `tfsdk:"tenant_id"`
	SubscriptionID    types.String              `tfsdk:"subscription_id"`
	ClientID          types.String              `tfsdk:"client_id"`
	ClientSecret      types.String              `tfsdk:"client_secret"`
	ServiceFilters    []azureServiceFilterModel `tfsdk:"service_filters"`
	ResourceGroups    types.List                `tfsdk:"resource_groups"`
	ResourceNameRegex types.String              `tfsdk:"resource_name_regex"`
}

// azureServiceFilterModel is one entry of the service_filters list. Both
// attributes are plain Optional/Required (never Optional+Computed), so the
// framework only ever plans them as null or known and a native slice would
// suffice; types.List/types.Map are used so an omitted value stays null
// instead of collapsing to [] and showing a permanent diff.
type azureServiceFilterModel struct {
	ServiceIDs types.List `tfsdk:"service_ids"`
	Tags       types.Map  `tfsdk:"tags"`
}

var _ resourceutils.ResourceModel[*clientmodels.AzureIntegration] = (*azureIntegrationResourceModel)(nil)

func (m *azureIntegrationResourceModel) GetID() types.String {
	return m.ID
}

func (m *azureIntegrationResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *azureIntegrationResourceModel) FromClientModel(
	ctx context.Context,
	model *clientmodels.AzureIntegration,
	diags *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.Status = types.StringValue(model.Status)

	az := model.TypeSpecificData.AzureMetricsIntegration
	m.SubscriptionName = types.StringValue(az.SubscriptionName)
	m.TenantID = types.StringValue(az.TenantID)
	m.SubscriptionID = types.StringValue(az.SubscriptionID)
	m.ClientID = types.StringValue(az.ClientID)

	// client_secret is write-only: every read returns AzureSecretMask rather
	// than the stored value, so the value already held is kept. Copying the
	// mask into state would otherwise show as permanent drift against the
	// configured secret. This relies on APIBaseResource passing the plan (not
	// a fresh model) as the receiver.

	m.ResourceGroups = resourceutils.SliceToStringList(
		ctx, az.ResourceGroups, m.ResourceGroups, diags,
	)
	m.ResourceNameRegex = optionalString(az.ResourceNameRegex, m.ResourceNameRegex)

	// Capture the prior filters before overwriting so each entry's optional
	// attributes can keep null rather than becoming empty collections.
	prior := m.ServiceFilters

	if len(az.ServiceFilters) == 0 {
		// The backend omits serviceFilters entirely when none are configured
		// (empty means "collect the default services"), so leave the unset
		// optional attribute null.
		m.ServiceFilters = nil
		return
	}

	filters := make([]azureServiceFilterModel, len(az.ServiceFilters))
	for i, filter := range az.ServiceFilters {
		var priorFilter azureServiceFilterModel
		if i < len(prior) {
			priorFilter = prior[i]
		}

		filters[i] = azureServiceFilterModel{
			ServiceIDs: resourceutils.SliceToStringList(
				ctx, filter.ServiceIDs, priorFilter.ServiceIDs, diags,
			),
			Tags: resourceutils.MapToStringMap(ctx, filter.Tags, priorFilter.Tags, diags),
		}
	}
	m.ServiceFilters = filters
}

func (m *azureIntegrationResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.AzureIntegration,
) error {
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		model.ID = m.ID.ValueString()
	}

	// Azure integrations are always written with the AZURE_METRICS
	// discriminator; the user does not set this on the resource.
	model.Type = clientmodels.AzureIntegrationType

	// name is Computed-only: the backend derives the integration name from
	// subscriptionName and ignores anything sent in the top-level field.

	az := clientmodels.AzureMetricsIntegration{
		SubscriptionName: m.SubscriptionName.ValueString(),
		TenantID:         m.TenantID.ValueString(),
		SubscriptionID:   m.SubscriptionID.ValueString(),
		ClientID:         m.ClientID.ValueString(),
		ClientSecret:     m.ClientSecret.ValueString(),
	}

	if !m.ResourceNameRegex.IsNull() && !m.ResourceNameRegex.IsUnknown() {
		az.ResourceNameRegex = m.ResourceNameRegex.ValueString()
	}

	resourceGroups, err := resourceutils.StringListToSlice(ctx, m.ResourceGroups)
	if err != nil {
		return fmt.Errorf("failed to read resource_groups: %w", err)
	}
	az.ResourceGroups = resourceGroups

	if len(m.ServiceFilters) > 0 {
		filters := make([]clientmodels.AzureServiceFilter, len(m.ServiceFilters))
		for i, filter := range m.ServiceFilters {
			serviceIDs, err := resourceutils.StringListToSlice(ctx, filter.ServiceIDs)
			if err != nil {
				return fmt.Errorf("failed to read service_ids for service_filters[%d]: %w", i, err)
			}

			entry := clientmodels.AzureServiceFilter{ServiceIDs: serviceIDs}
			if !filter.Tags.IsNull() && !filter.Tags.IsUnknown() {
				tags := map[string]string{}
				if d := filter.Tags.ElementsAs(ctx, &tags, false); d.HasError() {
					return fmt.Errorf("failed to read tags for service_filters[%d]: %v", i, d.Errors())
				}
				entry.Tags = tags
			}
			filters[i] = entry
		}
		az.ServiceFilters = filters
	}

	model.TypeSpecificData.AzureMetricsIntegration = az
	return nil
}

// optionalString keeps an optional attribute null when the backend returns no
// value, so omitting it in the configuration does not start showing a diff
// against an empty string.
//
// Unlike the same-named helpers in the GenAI resource packages, an empty value
// is only nulled when the prior value was null or unknown. resource_name_regex
// is Optional without Computed, so an explicitly configured "" has to stay a
// known empty string; collapsing it to null would fail the apply with
// "Provider produced inconsistent result after apply".
func optionalString(value string, prior types.String) types.String {
	if value == "" && (prior.IsNull() || prior.IsUnknown()) {
		return types.StringNull()
	}

	return types.StringValue(value)
}
