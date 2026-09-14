package gcpproject

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type gcpProjectResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Status                 types.String `tfsdk:"status"`
	Project                types.String `tfsdk:"project"`
	CustomerServiceAccount types.String `tfsdk:"customer_service_account"`
	OodlePrincipal         types.String `tfsdk:"oodle_principal"`
}

var _ resourceutils.ResourceModel[*clientmodels.GcpIntegration] = (*gcpProjectResourceModel)(nil)

func (m *gcpProjectResourceModel) GetID() types.String {
	return m.ID
}

func (m *gcpProjectResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *gcpProjectResourceModel) FromClientModel(
	_ context.Context,
	model *clientmodels.GcpIntegration,
	_ *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.Status = types.StringValue(model.Status)

	// A row this provider writes holds exactly one project. A row with none is
	// possible only when the project was removed outside Terraform; keep the
	// configured values so the next plan shows that as a diff to correct,
	// rather than emptying state and losing what the configuration asked for.
	if len(model.TypeSpecificData.Projects) == 0 {
		return
	}

	project := model.TypeSpecificData.Projects[0]
	m.Project = types.StringValue(project.Project)
	m.CustomerServiceAccount = types.StringValue(project.CustomerServiceAccount)
	m.OodlePrincipal = types.StringValue(project.OodlePrincipal)

	// The project carries the more precise status, but it stays empty until
	// the first status refresh runs. Keep the row status until then, so a
	// freshly created resource does not report an empty status.
	if project.Status != "" {
		m.Status = types.StringValue(project.Status)
	}
}

func (m *gcpProjectResourceModel) ToClientModel(
	_ context.Context,
	model *clientmodels.GcpIntegration,
) error {
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		model.ID = m.ID.ValueString()
	}

	// GCP integrations are always written with the GCP discriminator; the user
	// does not set this on the resource.
	model.Type = clientmodels.GcpIntegrationType

	// name is Computed-only: the backend names the row after its project and
	// ignores anything sent in the top-level field.

	if m.Project.IsNull() || m.Project.IsUnknown() {
		return fmt.Errorf("project is required")
	}

	// The backend rejects a project that carries no service account, because
	// the collector would have nothing to impersonate.
	if m.CustomerServiceAccount.IsNull() || m.CustomerServiceAccount.IsUnknown() {
		return fmt.Errorf("customer_service_account is required")
	}

	oodlePrincipal := clientmodels.DefaultOodlePrincipal
	if !m.OodlePrincipal.IsNull() && !m.OodlePrincipal.IsUnknown() {
		oodlePrincipal = m.OodlePrincipal.ValueString()
	}

	// One project per row is what gives the project an id of its own. Sending
	// more than one would recreate the shared array this resource exists to
	// replace.
	model.TypeSpecificData.Projects = []clientmodels.GcpProject{
		{
			Project:                m.Project.ValueString(),
			CustomerServiceAccount: m.CustomerServiceAccount.ValueString(),
			OodlePrincipal:         oodlePrincipal,
		},
	}

	return nil
}
