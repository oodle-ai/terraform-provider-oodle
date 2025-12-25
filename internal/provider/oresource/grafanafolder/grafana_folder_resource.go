package grafanafolder

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &grafanaFolderResource{}
	_ resource.ResourceWithConfigure   = &grafanaFolderResource{}
	_ resource.ResourceWithImportState = &grafanaFolderResource{}
)

type grafanaFolderResource struct {
	client *oodlehttp.GrafanaFolderClient
}

type grafanaFolderResourceModel struct {
	ID        types.String `tfsdk:"id"`
	UID       types.String `tfsdk:"uid"`
	Title     types.String `tfsdk:"title"`
	ParentUID types.String `tfsdk:"parent_uid"`
	URL       types.String `tfsdk:"url"`
}

func NewGrafanaFolderResource() resource.Resource {
	return &grafanaFolderResource{}
}

func (r *grafanaFolderResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_grafana_folder"
}

func (r *grafanaFolderResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a Grafana folder.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The UID of the folder (same as uid).",
			},
			"uid": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Unique identifier for the folder. " +
					"If not provided, one will be generated.",
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "The title of the folder.",
			},
			"parent_uid": schema.StringAttribute{
				Optional: true,
				Description: "The UID of the parent folder. " +
					"If not set, the folder will be created at the root level.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the folder in Grafana.",
			},
		},
	}
}

func (r *grafanaFolderResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*oodlehttp.OodleApiClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *oodlehttp.OodleApiClient, got: %T.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = oodlehttp.NewGrafanaFolderClient(client)
}

func (r *grafanaFolderResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan grafanaFolderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folder := &clientmodels.GrafanaFolder{
		Title:     plan.Title.ValueString(),
		UID:       plan.UID.ValueString(),
		ParentUID: plan.ParentUID.ValueString(),
	}

	created, err := r.client.Create(ctx, folder)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating folder",
			"Could not create folder: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(created.UID)
	plan.UID = types.StringValue(created.UID)
	plan.URL = types.StringValue(created.URL)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *grafanaFolderResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state grafanaFolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folder, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading folder",
			fmt.Sprintf(
				"Could not read folder %s: %v",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	state.ID = types.StringValue(folder.UID)
	state.UID = types.StringValue(folder.UID)
	state.Title = types.StringValue(folder.Title)
	state.URL = types.StringValue(folder.URL)
	if folder.ParentUID != "" {
		state.ParentUID = types.StringValue(folder.ParentUID)
	} else {
		state.ParentUID = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *grafanaFolderResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan grafanaFolderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state grafanaFolderResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	folder := &clientmodels.GrafanaFolder{
		UID:   state.ID.ValueString(),
		Title: plan.Title.ValueString(),
	}

	updated, err := r.client.Update(ctx, folder)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating folder",
			fmt.Sprintf(
				"Could not update folder %s: %v",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	plan.ID = types.StringValue(updated.UID)
	plan.UID = types.StringValue(updated.UID)
	plan.URL = types.StringValue(updated.URL)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *grafanaFolderResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state grafanaFolderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting folder",
			fmt.Sprintf(
				"Could not delete folder %s: %v",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}
}

func (r *grafanaFolderResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
