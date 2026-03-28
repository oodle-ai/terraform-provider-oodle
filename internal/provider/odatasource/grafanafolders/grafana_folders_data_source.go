package grafanafolders

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
	_ datasource.DataSource              = &grafanaFoldersDataSource{}
	_ datasource.DataSourceWithConfigure = &grafanaFoldersDataSource{}
)

type grafanaFoldersDataSource struct {
	client *oodlehttp.GrafanaFolderClient
}

type grafanaFoldersDataSourceModel struct {
	Folders []folderModel `tfsdk:"folders"`
}

type folderModel struct {
	UID   types.String `tfsdk:"uid"`
	Title types.String `tfsdk:"title"`
}

func NewGrafanaFoldersDataSource() datasource.DataSource {
	return &grafanaFoldersDataSource{}
}

func (d *grafanaFoldersDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_grafana_folders"
}

func (d *grafanaFoldersDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Lists all Grafana folders.",
		Attributes: map[string]schema.Attribute{
			"folders": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Grafana folders.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uid": schema.StringAttribute{
							Computed:    true,
							Description: "The UID of the folder.",
						},
						"title": schema.StringAttribute{
							Computed:    true,
							Description: "The title of the folder.",
						},
					},
				},
			},
		},
	}
}

func (d *grafanaFoldersDataSource) Configure(
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

	d.client = oodlehttp.NewGrafanaFolderClient(client)
}

func (d *grafanaFoldersDataSource) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	folders, err := d.client.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing Grafana folders",
			"Could not list Grafana folders: "+err.Error(),
		)
		return
	}

	state := grafanaFoldersDataSourceModel{
		Folders: make([]folderModel, 0, len(folders)),
	}
	for _, f := range folders {
		state.Folders = append(state.Folders, folderModel{
			UID:   types.StringValue(f.UID),
			Title: types.StringValue(f.Title),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
