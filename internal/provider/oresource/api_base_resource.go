package oresource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

// ModelAPI is the CRUD surface a GenAI object's client exposes.
// The id passed to Get and Delete is whatever the resource stores in
// its "id" attribute, which for datasets is the name rather than a
// uuid — those endpoints are keyed by name.
type ModelAPI[M clientmodels.ClientModel] interface {
	Create(ctx context.Context, model M) (M, error)
	Get(ctx context.Context, id string) (M, error)
	Update(ctx context.Context, model M) (M, error)
	Delete(ctx context.Context, id string) error
}

// APIBaseResource wires a GenAI object's client to the Terraform
// resource lifecycle.
//
// It differs from BaseResource in two ways that matter for these
// APIs. A read that 404s removes the resource from state instead of
// failing the plan, so an object deleted in the UI is recreated
// rather than wedging every subsequent apply. And the plan is used as
// the receiver when converting a response back into state, so that
// attributes the API never echoes — an LLM connection's api_key —
// keep their configured value instead of reverting to null.
type APIBaseResource[M clientmodels.ClientModel, R resourceutils.ResourceModel[M]] struct {
	client           ModelAPI[M]
	newResourceModel func() R
	newClientModel   func() M
	createClient     func(oodleHttpClient *oodlehttp.OodleApiClient) ModelAPI[M]
}

func NewAPIBaseResource[M clientmodels.ClientModel, R resourceutils.ResourceModel[M]](
	newResourceModel func() R,
	newClientModel func() M,
	createClient func(oodleHttpClient *oodlehttp.OodleApiClient) ModelAPI[M],
) APIBaseResource[M, R] {
	return APIBaseResource[M, R]{
		newResourceModel: newResourceModel,
		newClientModel:   newClientModel,
		createClient:     createClient,
	}
}

// Configure adds the provider configured client to the resource.
func (r *APIBaseResource[M, R]) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*oodlehttp.OodleApiClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *oodlehttp.OodleApiClient, got: %T. "+
					"Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	r.client = r.createClient(client)
}

func (r *APIBaseResource[M, R]) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a new resource.
func (r *APIBaseResource[M, R]) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	plan := r.newResourceModel()
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientModel := r.newClientModel()
	if err := plan.ToClientModel(ctx, clientModel); err != nil {
		resp.Diagnostics.AddError("Failed to convert plan to model", err.Error())
		return
	}

	createdObj, err := r.client.Create(ctx, clientModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating model",
			"Could not create model, unexpected error: "+err.Error(),
		)
		return
	}

	plan.FromClientModel(ctx, createdObj, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *APIBaseResource[M, R]) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	state := r.newResourceModel()
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.GetID()
	if id.IsNull() || id.IsUnknown() {
		resp.Diagnostics.AddError(
			"ID is not set",
			fmt.Sprintf("ID is required to read obj: %+v", state),
		)
		return
	}

	obj, err := r.client.Get(ctx, id.ValueString())
	if err != nil {
		if errors.Is(err, oodlehttp.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading model",
			fmt.Sprintf("Could not read %q: %v", id.ValueString(), err),
		)
		return
	}

	state.FromClientModel(ctx, obj, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *APIBaseResource[M, R]) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	plan := r.newResourceModel()
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The id is server-assigned, so it lives in state rather than
	// in the plan.
	state := r.newResourceModel()
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.SetID(state.GetID())

	id := plan.GetID()
	if id.IsNull() || id.IsUnknown() {
		resp.Diagnostics.AddError(
			"ID is not set",
			fmt.Sprintf("ID is required to update obj: %+v", plan),
		)
		return
	}

	model := r.newClientModel()
	if err := plan.ToClientModel(ctx, model); err != nil {
		resp.Diagnostics.AddError("Failed to convert plan to model", err.Error())
		return
	}

	updatedObj, err := r.client.Update(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating model",
			fmt.Sprintf(
				"Could not update %q, unexpected error: %v",
				id.ValueString(),
				err,
			),
		)
		return
	}

	plan.FromClientModel(ctx, updatedObj, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *APIBaseResource[M, R]) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	state := r.newResourceModel()
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.GetID()
	if id.IsNull() || id.IsUnknown() {
		resp.Diagnostics.AddError(
			"ID is not set",
			fmt.Sprintf("ID is required to delete obj: %+v", state),
		)
		return
	}

	err := r.client.Delete(ctx, id.ValueString())
	if err != nil {
		// Already gone is the desired end state.
		if errors.Is(err, oodlehttp.ErrNotFound) {
			return
		}

		resp.Diagnostics.AddError(
			"Error deleting model",
			fmt.Sprintf("Could not delete %q: %v", id.ValueString(), err),
		)
		return
	}
}
