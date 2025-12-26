package grafanadashboard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	jsoniter "github.com/json-iterator/go"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &grafanaDashboardResource{}
	_ resource.ResourceWithConfigure   = &grafanaDashboardResource{}
	_ resource.ResourceWithImportState = &grafanaDashboardResource{}
)

type grafanaDashboardResource struct {
	client *oodlehttp.GrafanaDashboardClient
}

type grafanaDashboardResourceModel struct {
	ID         types.String `tfsdk:"id"`
	UID        types.String `tfsdk:"uid"`
	ConfigJSON types.String `tfsdk:"config_json"`
	Folder     types.String `tfsdk:"folder"`
	Overwrite  types.Bool   `tfsdk:"overwrite"`
	Message    types.String `tfsdk:"message"`
	URL        types.String `tfsdk:"url"`
	Version    types.Int64  `tfsdk:"version"`
}

func NewGrafanaDashboardResource() resource.Resource {
	return &grafanaDashboardResource{}
}

func (r *grafanaDashboardResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_grafana_dashboard"
}

func (r *grafanaDashboardResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a Grafana dashboard.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The UID of the dashboard.",
			},
			"uid": schema.StringAttribute{
				Computed:    true,
				Description: "The UID of the dashboard (same as id).",
			},
			"config_json": schema.StringAttribute{
				Required: true,
				Description: "The complete dashboard model JSON. " +
					"The 'uid' field in the JSON will be used as the dashboard identifier.",
			},
			"folder": schema.StringAttribute{
				Optional:    true,
				Description: "The UID of the folder to save the dashboard in.",
			},
			"overwrite": schema.BoolAttribute{
				Optional: true,
				Description: "Set to true if you want to overwrite existing dashboard " +
					"with newer version or with same dashboard title.",
			},
			"message": schema.StringAttribute{
				Optional:    true,
				Description: "Set a commit message for the version history.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the dashboard in Grafana.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The version of the dashboard.",
			},
		},
	}
}

func (r *grafanaDashboardResource) Configure(
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

	r.client = oodlehttp.NewGrafanaDashboardClient(client)
}

func (r *grafanaDashboardResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan grafanaDashboardResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var dashboardData interface{}
	if err := jsoniter.Unmarshal(
		[]byte(plan.ConfigJSON.ValueString()),
		&dashboardData,
	); err != nil {
		resp.Diagnostics.AddError(
			"Invalid config_json",
			"Could not parse config_json: "+err.Error(),
		)
		return
	}

	dashboard := &clientmodels.GrafanaDashboard{
		Dashboard: dashboardData,
		FolderUID: plan.Folder.ValueString(),
		Overwrite: plan.Overwrite.ValueBool(),
		Message:   plan.Message.ValueString(),
	}

	created, err := r.client.Create(ctx, dashboard)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating dashboard",
			"Could not create dashboard: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(created.UID)
	plan.UID = types.StringValue(created.UID)
	plan.URL = types.StringValue(created.URL)
	plan.Version = types.Int64Value(int64(created.Version))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *grafanaDashboardResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state grafanaDashboardResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboard, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading dashboard",
			fmt.Sprintf(
				"Could not read dashboard %s: %v",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	configJSON, err := dashboard.GetConfigJSON()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error serializing dashboard",
			"Could not serialize dashboard JSON: "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(dashboard.GetID())
	state.UID = types.StringValue(dashboard.GetID())
	state.ConfigJSON = types.StringValue(configJSON)
	state.URL = types.StringValue(dashboard.Meta.URL)
	state.Version = types.Int64Value(int64(dashboard.Meta.Version))
	if dashboard.Meta.FolderUID != "" {
		state.Folder = types.StringValue(dashboard.Meta.FolderUID)
	} else {
		state.Folder = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *grafanaDashboardResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan grafanaDashboardResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state grafanaDashboardResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var dashboardData interface{}
	if err := jsoniter.Unmarshal(
		[]byte(plan.ConfigJSON.ValueString()),
		&dashboardData,
	); err != nil {
		resp.Diagnostics.AddError(
			"Invalid config_json",
			"Could not parse config_json: "+err.Error(),
		)
		return
	}

	// Ensure UID matches the state
	if dashMap, ok := dashboardData.(map[string]interface{}); ok {
		dashMap["uid"] = state.ID.ValueString()
	}

	dashboard := &clientmodels.GrafanaDashboard{
		Dashboard: dashboardData,
		FolderUID: plan.Folder.ValueString(),
		Overwrite: true,
		Message:   plan.Message.ValueString(),
	}

	updated, err := r.client.Update(ctx, dashboard)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating dashboard",
			fmt.Sprintf(
				"Could not update dashboard %s: %v",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	plan.ID = types.StringValue(updated.UID)
	plan.UID = types.StringValue(updated.UID)
	plan.URL = types.StringValue(updated.URL)
	plan.Version = types.Int64Value(int64(updated.Version))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *grafanaDashboardResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state grafanaDashboardResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting dashboard",
			fmt.Sprintf(
				"Could not delete dashboard %s: %v",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}
}

func (r *grafanaDashboardResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
