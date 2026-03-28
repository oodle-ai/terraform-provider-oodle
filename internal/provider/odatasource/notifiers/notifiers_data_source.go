package notifiers

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
	_ datasource.DataSource              = &notifiersDataSource{}
	_ datasource.DataSourceWithConfigure = &notifiersDataSource{}
)

type notifiersDataSource struct {
	client *oodlehttp.ModelClient[*clientmodels.Notifier]
}

type notifiersDataSourceModel struct {
	Notifiers []notifierModel `tfsdk:"notifiers"`
}

type notifierModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

func NewNotifiersDataSource() datasource.DataSource {
	return &notifiersDataSource{}
}

func (d *notifiersDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notifiers"
}

func (d *notifiersDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Lists all notifiers.",
		Attributes: map[string]schema.Attribute{
			"notifiers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of notifiers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the notifier.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the notifier.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the notifier (e.g. email, pagerduty, slack, opsgenie, webhook, googlechat).",
						},
					},
				},
			},
		},
	}
}

func (d *notifiersDataSource) Configure(
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

	d.client = oodlehttp.NewModelClient[*clientmodels.Notifier](
		client,
		"notifiers",
		func() *clientmodels.Notifier { return &clientmodels.Notifier{} },
	)
}

func (d *notifiersDataSource) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	notifiersList, err := d.client.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing notifiers",
			"Could not list notifiers: "+err.Error(),
		)
		return
	}

	state := notifiersDataSourceModel{
		Notifiers: make([]notifierModel, 0, len(notifiersList)),
	}
	for _, n := range notifiersList {
		state.Notifiers = append(state.Notifiers, notifierModel{
			ID:   types.StringValue(n.GetID()),
			Name: types.StringValue(n.Name),
			Type: types.StringValue(n.Type.String()),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
