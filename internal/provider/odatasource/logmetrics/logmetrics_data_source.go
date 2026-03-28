package logmetrics

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
	_ datasource.DataSource              = &logmetricsDataSource{}
	_ datasource.DataSourceWithConfigure = &logmetricsDataSource{}
)

type logmetricsDataSource struct {
	client *oodlehttp.ModelClient[*clientmodels.LogMetrics]
}

type logmetricsDataSourceModel struct {
	Logmetrics []logmetricsModel `tfsdk:"logmetrics"`
}

type logmetricsModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewLogmetricsDataSource() datasource.DataSource {
	return &logmetricsDataSource{}
}

func (d *logmetricsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_logmetrics"
}

func (d *logmetricsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Lists all log metrics rules.",
		Attributes: map[string]schema.Attribute{
			"logmetrics": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of log metrics rules.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the log metrics rule.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the log metrics rule.",
						},
					},
				},
			},
		},
	}
}

func (d *logmetricsDataSource) Configure(
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

	d.client = oodlehttp.NewModelClient[*clientmodels.LogMetrics](
		client,
		"logmetrics",
		func() *clientmodels.LogMetrics { return &clientmodels.LogMetrics{} },
	)
}

func (d *logmetricsDataSource) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	logmetricsList, err := d.client.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing log metrics",
			"Could not list log metrics: "+err.Error(),
		)
		return
	}

	state := logmetricsDataSourceModel{
		Logmetrics: make([]logmetricsModel, 0, len(logmetricsList)),
	}
	for _, l := range logmetricsList {
		state.Logmetrics = append(state.Logmetrics, logmetricsModel{
			ID:   types.StringValue(l.GetID()),
			Name: types.StringValue(l.Name),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
