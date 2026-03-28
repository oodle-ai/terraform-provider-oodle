package monitors

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &monitorsDataSource{}
	_ datasource.DataSourceWithConfigure = &monitorsDataSource{}
)

type monitorsDataSource struct {
	client *oodlehttp.ModelClient[*clientmodels.Monitor]
}

type monitorsDataSourceModel struct {
	Monitors []monitorModel `tfsdk:"monitors"`
}

type monitorModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewMonitorsDataSource() datasource.DataSource {
	return &monitorsDataSource{}
}

func (d *monitorsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

func (d *monitorsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Lists all monitors.",
		Attributes: map[string]schema.Attribute{
			"monitors": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of monitors.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the monitor.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the monitor.",
						},
					},
				},
			},
		},
	}
}

func (d *monitorsDataSource) Configure(
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

	d.client = oodlehttp.NewModelClient[*clientmodels.Monitor](
		client,
		"monitors",
		func() *clientmodels.Monitor { return &clientmodels.Monitor{} },
	)
}

func (d *monitorsDataSource) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	monitors, err := d.client.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing monitors",
			"Could not list monitors: "+err.Error(),
		)
		return
	}

	state := monitorsDataSourceModel{
		Monitors: make([]monitorModel, 0, len(monitors)),
	}
	for _, m := range monitors {
		state.Monitors = append(state.Monitors, monitorModel{
			ID:   types.StringValue(m.GetID()),
			Name: types.StringValue(m.Name),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
