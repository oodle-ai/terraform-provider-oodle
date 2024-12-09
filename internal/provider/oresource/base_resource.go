package oresource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type BaseResource[M clientmodels.ClientModel, R resourceutils.ResourceModel[M]] struct {
	client           *oodlehttp.ModelClient[M]
	newResourceModel func() R
	newClientModel   func() M
	createClient     func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[M]
}

func NewBaseResource[M clientmodels.ClientModel, R resourceutils.ResourceModel[M]](
	newResourceModel func() R,
	newClientModel func() M,
	createClient func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[M],
) BaseResource[M, R] {
	return BaseResource[M, R]{
		newResourceModel: newResourceModel,
		newClientModel:   newClientModel,
		createClient:     createClient,
	}
}

// Configure adds the provider configured client to the resource.
func (r *BaseResource[M, R]) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*oodlehttp.OodleApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *oodlehttp.OodleApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = r.createClient(client)
}

func (r *BaseResource[M, R]) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a new resource.
func (r *BaseResource[M, R]) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	plan := r.newResourceModel()
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientModel := r.newClientModel()
	err := plan.ToClientModel(clientModel)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert plan to model", err.Error())
		return
	}

	// Create new monitor
	createdMonitor, err := r.client.Create(clientModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Model",
			"Could not create model, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with newly created monitor.
	newPlan := r.newResourceModel()
	newPlan.FromClientModel(createdMonitor, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newPlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *BaseResource[M, R]) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := r.newResourceModel()
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.GetID()
	if id.IsNull() || id.IsUnknown() {
		resp.Diagnostics.AddError("ID is not set", fmt.Sprintf("ID is required to read monitor: %+v", state))
		return
	}

	// Get refreshed order value from HashiCups
	monitor, err := r.client.Get(id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Monitors",
			"Could not read Oodle Monitor ID "+id.ValueString()+": "+err.Error(),
		)
		return
	}

	state.FromClientModel(monitor, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *BaseResource[M, R]) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	plan := r.newResourceModel()
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Assign ID to plan from state.
	state := r.newResourceModel()
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.SetID(state.GetID())

	id := plan.GetID()
	if id.IsNull() || id.IsUnknown() {
		resp.Diagnostics.AddError("ID is not set", fmt.Sprintf("ID is required to update monitor: %+v", plan))
		return
	}

	model := r.newClientModel()
	err := plan.ToClientModel(model)
	if err != nil {
		resp.Diagnostics.AddError("Failed to convert plan to model", err.Error())
		return
	}

	updatedMonitor, err := r.client.Update(model)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Monitor",
			"Could not update monitor, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with newly created monitor.
	newState := r.newResourceModel()
	newState.FromClientModel(updatedMonitor, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *BaseResource[M, R]) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	state := r.newResourceModel()
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.GetID()
	if id.IsNull() || id.IsUnknown() {
		resp.Diagnostics.AddError("ID is not set", fmt.Sprintf("ID is required to delete monitor: %+v", state))
		return
	}

	// Delete existing order
	err := r.client.Delete(id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Monitor",
			fmt.Sprintf("Could not delete monitor ID %s: %v", id.ValueString(), err),
		)
		return
	}
}
