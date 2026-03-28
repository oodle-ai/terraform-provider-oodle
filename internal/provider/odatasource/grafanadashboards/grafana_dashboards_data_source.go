package grafanadashboards

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &grafanaDashboardsDataSource{}
	_ datasource.DataSourceWithConfigure = &grafanaDashboardsDataSource{}
)

type grafanaDashboardsDataSource struct {
	client *oodlehttp.GrafanaDashboardClient
}

type grafanaDashboardsDataSourceModel struct {
	Dashboards []dashboardModel `tfsdk:"dashboards"`
}

type dashboardModel struct {
	UID         types.String `tfsdk:"uid"`
	Title       types.String `tfsdk:"title"`
	FolderUID   types.String `tfsdk:"folder_uid"`
	FolderTitle types.String `tfsdk:"folder_title"`
	URL         types.String `tfsdk:"url"`
	Type        types.String `tfsdk:"type"`
}

func NewGrafanaDashboardsDataSource() datasource.DataSource {
	return &grafanaDashboardsDataSource{}
}

func (d *grafanaDashboardsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_grafana_dashboards"
}

func (d *grafanaDashboardsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Lists all Grafana dashboards.",
		Attributes: map[string]schema.Attribute{
			"dashboards": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Grafana dashboards.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uid": schema.StringAttribute{
							Computed:    true,
							Description: "The UID of the dashboard.",
						},
						"title": schema.StringAttribute{
							Computed:    true,
							Description: "The title of the dashboard.",
						},
						"folder_uid": schema.StringAttribute{
							Computed:    true,
							Description: "The UID of the folder containing the dashboard.",
						},
						"folder_title": schema.StringAttribute{
							Computed:    true,
							Description: "The title of the folder containing the dashboard.",
						},
						"url": schema.StringAttribute{
							Computed:    true,
							Description: "The URL of the dashboard.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the dashboard entry (e.g. dash-db, dash-folder).",
						},
					},
				},
			},
		},
	}
}

func (d *grafanaDashboardsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*oodlehttp.OodleApiClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *oodlehttp.OodleApiClient, got: %T.",
				req.ProviderData,
			),
		)
		return
	}

	d.client = oodlehttp.NewGrafanaDashboardClient(client)
}

func (d *grafanaDashboardsDataSource) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	dashboards, err := d.client.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing Grafana dashboards",
			"Could not list Grafana dashboards: "+err.Error(),
		)
		return
	}

	state := grafanaDashboardsDataSourceModel{
		Dashboards: make([]dashboardModel, 0, len(dashboards)),
	}
	for _, db := range dashboards {
		state.Dashboards = append(state.Dashboards, dashboardModel{
			UID:         types.StringValue(db.UID),
			Title:       types.StringValue(db.Title),
			FolderUID:   types.StringValue(db.FolderUID),
			FolderTitle: types.StringValue(db.FolderTitle),
			URL:         types.StringValue(db.URL),
			Type:        types.StringValue(db.Type),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
