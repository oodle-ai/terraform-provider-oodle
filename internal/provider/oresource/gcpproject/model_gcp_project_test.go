package gcpproject

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestGcpProjectModelRoundTrip(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.GcpIntegration{
		ID:     "01a08f55-e42b-76fa-a1ac-f60f26363260",
		Type:   clientmodels.GcpIntegrationType,
		Name:   "acme-prod",
		Status: "RECEIVING",
	}
	clientModel.TypeSpecificData.Projects = []clientmodels.GcpProject{
		{
			Status:                 "RECEIVING",
			Project:                "acme-prod",
			CustomerServiceAccount: "oodle@acme-prod.iam.gserviceaccount.com",
			OodlePrincipal:         clientmodels.DefaultOodlePrincipal,
		},
	}

	resourceModel := &gcpProjectResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.GcpIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	// Everything round-trips except the fields the server owns: the top-level
	// name (derived from the project), both statuses, and the principal, which
	// the server fills with the one it uses for every customer.
	expected := *clientModel
	expected.Name = ""
	expected.Status = ""
	expected.TypeSpecificData.Projects[0].Status = ""
	expected.TypeSpecificData.Projects[0].OodlePrincipal = ""
	assert.DeepEqual(t, &expected, newClientModel)
}

// TestGcpProjectModelWritesOneProject is the guard for the whole point of the
// resource. A row that holds several projects is the shared array this
// resource replaces: Terraform cannot address one project of it, and a write
// would overwrite its neighbours.
func TestGcpProjectModelWritesOneProject(t *testing.T) {
	ctx := context.Background()
	resourceModel := &gcpProjectResourceModel{
		Project:                types.StringValue("acme-prod"),
		CustomerServiceAccount: types.StringValue("oodle@acme-prod.iam.gserviceaccount.com"),
	}

	clientModel := &clientmodels.GcpIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, clientModel))
	assert.Equal(t, 1, len(clientModel.TypeSpecificData.Projects))
	assert.Equal(t, clientmodels.GcpIntegrationType, clientModel.Type)
}

// TestGcpProjectModelKeepsConfigOnEmptyRead covers a project removed outside
// Terraform. The row still exists, so the read is not a 404 and the resource
// stays in state; emptying the configured attributes would report the drift as
// "no change needed" instead of a diff to correct.
func TestGcpProjectModelKeepsConfigOnEmptyRead(t *testing.T) {
	ctx := context.Background()
	resourceModel := &gcpProjectResourceModel{
		Project:                types.StringValue("acme-prod"),
		CustomerServiceAccount: types.StringValue("oodle@acme-prod.iam.gserviceaccount.com"),
	}

	clientModel := &clientmodels.GcpIntegration{
		ID:   "01a08f55-e42b-76fa-a1ac-f60f26363260",
		Type: clientmodels.GcpIntegrationType,
	}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.Equal(t, "acme-prod", resourceModel.Project.ValueString())
	assert.Equal(
		t,
		"oodle@acme-prod.iam.gserviceaccount.com",
		resourceModel.CustomerServiceAccount.ValueString(),
	)
}

// TestGcpProjectModelRequiresServiceAccount pins the client-side check. The
// backend rejects a project with no service account, and catching it here
// turns an apply-time error into one the user sees before any write.
func TestGcpProjectModelRequiresServiceAccount(t *testing.T) {
	ctx := context.Background()
	resourceModel := &gcpProjectResourceModel{
		Project: types.StringValue("acme-prod"),
	}

	assert.NotNil(t, resourceModel.ToClientModel(ctx, &clientmodels.GcpIntegration{}))
}

// TestGcpProjectModelPrefersRowStatusBeforeFirstRefresh covers a project
// created moments ago. The per-project status is written by the status
// refresher, so until it first runs the row status is the only one there is.
func TestGcpProjectModelPrefersRowStatusBeforeFirstRefresh(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.GcpIntegration{
		ID:     "01a08f55-e42b-76fa-a1ac-f60f26363260",
		Type:   clientmodels.GcpIntegrationType,
		Status: "NOT_CONNECTED",
	}
	clientModel.TypeSpecificData.Projects = []clientmodels.GcpProject{
		{
			Project:                "acme-prod",
			CustomerServiceAccount: "oodle@acme-prod.iam.gserviceaccount.com",
		},
	}

	resourceModel := &gcpProjectResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())
	assert.Equal(t, "NOT_CONNECTED", resourceModel.Status.ValueString())
}
