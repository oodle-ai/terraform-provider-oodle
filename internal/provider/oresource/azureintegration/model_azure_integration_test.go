package azureintegration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestAzureIntegrationModelRoundTrip(t *testing.T) {
	ctx := context.Background()
	// Reads always carry the masked secret; the plan that APIBaseResource
	// passes as the receiver carries the configured one.
	clientModel := &clientmodels.AzureIntegration{
		ID:   "01a08f55-e42b-76fa-a1ac-f60f26363260",
		Type: clientmodels.AzureIntegrationType,
		Name: "OodleAI POC",
	}
	clientModel.TypeSpecificData.AzureMetricsIntegration = clientmodels.AzureMetricsIntegration{
		SubscriptionName: "OodleAI POC",
		TenantID:         "9323ac63-de0e-4f40-93f4-20ea0a7a3b19",
		SubscriptionID:   "515daa61-af2c-408d-8416-8d265535cf5c",
		ClientID:         "a3bce0d5-01d4-48b2-9d5f-b1d1368852ac",
		ClientSecret:     clientmodels.AzureSecretMask,
		ServiceFilters: []clientmodels.AzureServiceFilter{
			{
				ServiceIDs: []string{"compute_virtualmachines", "storage_storageaccounts"},
				Tags:       map[string]string{"environment": "production"},
			},
			{ServiceIDs: []string{"cache_redis"}},
		},
		ResourceGroups:    []string{"rg-prod-east"},
		ResourceNameRegex: "^prod-",
	}

	resourceModel := &azureIntegrationResourceModel{
		ClientSecret: types.StringValue("configured-secret"),
	}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.AzureIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	// Everything round-trips unchanged except two deliberate omissions: the
	// secret goes back as the configured value rather than the mask that was
	// read, and the top-level name is dropped because the backend derives it
	// from subscriptionName and ignores whatever is sent.
	expected := *clientModel
	expected.Name = ""
	expected.TypeSpecificData.AzureMetricsIntegration.ClientSecret = "configured-secret"
	assert.DeepEqual(t, &expected, newClientModel)
}

// TestAzureIntegrationModelSecretMaskNotStored is the guard for the write-only
// secret. The backend answers every read with "********", and copying that into
// state would make every subsequent plan show a diff against the configured
// secret (and would eventually send the mask back as if it were the real
// value). FromClientModel must leave the attribute exactly as the plan had it.
func TestAzureIntegrationModelSecretMaskNotStored(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.AzureIntegration{
		ID:   "int-secret",
		Type: clientmodels.AzureIntegrationType,
	}
	clientModel.TypeSpecificData.AzureMetricsIntegration = clientmodels.AzureMetricsIntegration{
		SubscriptionName: "OodleAI POC",
		TenantID:         "tenant-id",
		SubscriptionID:   "subscription-id",
		ClientID:         "client-id",
		ClientSecret:     clientmodels.AzureSecretMask,
	}

	resourceModel := &azureIntegrationResourceModel{
		ClientSecret: types.StringValue("configured-secret"),
	}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.Equal(t, "configured-secret", resourceModel.ClientSecret.ValueString())

	echoed := &clientmodels.AzureIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, echoed))
	assert.Equal(
		t,
		"configured-secret",
		echoed.TypeSpecificData.AzureMetricsIntegration.ClientSecret,
	)
}

// TestAzureIntegrationModelFromServer verifies that the Computed-only fields
// land in state, and that name is not echoed back: the backend derives the
// integration name from subscriptionName and ignores the top-level field.
func TestAzureIntegrationModelFromServer(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.AzureIntegration{
		ID:     "int-server-set",
		Type:   clientmodels.AzureIntegrationType,
		Name:   "OodleAI POC",
		Status: "RECEIVING",
	}
	clientModel.TypeSpecificData.AzureMetricsIntegration = clientmodels.AzureMetricsIntegration{
		SubscriptionName: "OodleAI POC",
		TenantID:         "tenant-id",
		SubscriptionID:   "subscription-id",
		ClientID:         "client-id",
		ClientSecret:     clientmodels.AzureSecretMask,
	}

	resourceModel := &azureIntegrationResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.Equal(t, "RECEIVING", resourceModel.Status.ValueString())
	assert.Equal(t, "OodleAI POC", resourceModel.Name.ValueString())
	assert.Equal(t, "int-server-set", resourceModel.ID.ValueString())

	echoed := &clientmodels.AzureIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, echoed))
	assert.Equal(t, "", echoed.Status)
	assert.Equal(t, "", echoed.Name)
	assert.Equal(t, clientmodels.AzureIntegrationType, echoed.Type)
}

// TestAzureIntegrationModelOmittedOptionalsStayNull guards against a permanent
// diff on the minimal configuration. The backend omits serviceFilters,
// resourceGroups and resourceNameRegex from the response when they are unset,
// and turning those into [] / {} / "" in state would differ from the null the
// plan holds.
func TestAzureIntegrationModelOmittedOptionalsStayNull(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.AzureIntegration{
		ID:   "int-minimal",
		Type: clientmodels.AzureIntegrationType,
		Name: "minimal",
	}
	clientModel.TypeSpecificData.AzureMetricsIntegration = clientmodels.AzureMetricsIntegration{
		SubscriptionName: "minimal",
		TenantID:         "tenant-id",
		SubscriptionID:   "subscription-id",
		ClientID:         "client-id",
		ClientSecret:     clientmodels.AzureSecretMask,
	}

	resourceModel := &azureIntegrationResourceModel{
		ClientSecret: types.StringValue("configured-secret"),
	}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.Nil(t, resourceModel.ServiceFilters)
	assert.True(t, resourceModel.ResourceGroups.IsNull())
	assert.True(t, resourceModel.ResourceNameRegex.IsNull())

	// And the minimal shape survives the trip back without inventing empty
	// collections that the backend would then store.
	newClientModel := &clientmodels.AzureIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	az := newClientModel.TypeSpecificData.AzureMetricsIntegration
	assert.Nil(t, az.ServiceFilters)
	assert.Nil(t, az.ResourceGroups)
	assert.Equal(t, "", az.ResourceNameRegex)
}

// TestAzureIntegrationModelOmittedTagsStayNull covers the same null-vs-empty
// concern one level down: a service filter group with no tag filter must not
// acquire an empty map in state.
func TestAzureIntegrationModelOmittedTagsStayNull(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.AzureIntegration{
		ID:   "int-no-tags",
		Type: clientmodels.AzureIntegrationType,
	}
	clientModel.TypeSpecificData.AzureMetricsIntegration = clientmodels.AzureMetricsIntegration{
		SubscriptionName: "no-tags",
		TenantID:         "tenant-id",
		SubscriptionID:   "subscription-id",
		ClientID:         "client-id",
		ClientSecret:     clientmodels.AzureSecretMask,
		ServiceFilters: []clientmodels.AzureServiceFilter{
			{ServiceIDs: []string{"compute_virtualmachines"}},
		},
	}

	resourceModel := &azureIntegrationResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.Equal(t, 1, len(resourceModel.ServiceFilters))
	assert.True(t, resourceModel.ServiceFilters[0].Tags.IsNull())
	assert.False(t, resourceModel.ServiceFilters[0].ServiceIDs.IsNull())
}

// TestAzureIntegrationModelExplicitEmptyTagsStayKnown is the counterpart: a
// config that writes `tags = {}` must keep a known empty map, or Create fails
// with "Provider produced inconsistent result after apply" because the applied
// state (null) would differ from the plan ({}).
func TestAzureIntegrationModelExplicitEmptyTagsStayKnown(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.AzureIntegration{
		ID:   "int-empty-tags",
		Type: clientmodels.AzureIntegrationType,
	}
	clientModel.TypeSpecificData.AzureMetricsIntegration = clientmodels.AzureMetricsIntegration{
		SubscriptionName: "empty-tags",
		TenantID:         "tenant-id",
		SubscriptionID:   "subscription-id",
		ClientID:         "client-id",
		ClientSecret:     clientmodels.AzureSecretMask,
		ServiceFilters: []clientmodels.AzureServiceFilter{
			{ServiceIDs: []string{"compute_virtualmachines"}},
		},
	}

	resourceModel := &azureIntegrationResourceModel{
		ServiceFilters: []azureServiceFilterModel{
			{
				ServiceIDs: types.ListValueMust(
					types.StringType,
					[]attr.Value{types.StringValue("compute_virtualmachines")},
				),
				Tags: types.MapValueMust(types.StringType, map[string]attr.Value{}),
			},
		},
	}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.Equal(t, 1, len(resourceModel.ServiceFilters))
	assert.False(t, resourceModel.ServiceFilters[0].Tags.IsNull())
	assert.Equal(t, 0, len(resourceModel.ServiceFilters[0].Tags.Elements()))
}

// TestAzureIntegrationModelImportedSecretIsNull documents the import path:
// there is no configuration to preserve, so the write-only secret lands null
// in state and the first plan after an import shows a diff until the secret is
// declared. Asserted so the behaviour is a decision rather than a surprise.
func TestAzureIntegrationModelImportedSecretIsNull(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.AzureIntegration{
		ID:   "int-imported",
		Type: clientmodels.AzureIntegrationType,
	}
	clientModel.TypeSpecificData.AzureMetricsIntegration = clientmodels.AzureMetricsIntegration{
		SubscriptionName: "imported",
		TenantID:         "tenant-id",
		SubscriptionID:   "subscription-id",
		ClientID:         "client-id",
		ClientSecret:     clientmodels.AzureSecretMask,
	}

	resourceModel := &azureIntegrationResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.True(t, resourceModel.ClientSecret.IsNull())
}
