package notificationpolicies

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
	_ datasource.DataSource              = &notificationPoliciesDataSource{}
	_ datasource.DataSourceWithConfigure = &notificationPoliciesDataSource{}
)

type notificationPoliciesDataSource struct {
	client *oodlehttp.ModelClient[*clientmodels.NotificationPolicy]
}

type notificationPoliciesDataSourceModel struct {
	NotificationPolicies []notificationPolicyModel `tfsdk:"notification_policies"`
}

type notificationPolicyModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewNotificationPoliciesDataSource() datasource.DataSource {
	return &notificationPoliciesDataSource{}
}

func (d *notificationPoliciesDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notification_policies"
}

func (d *notificationPoliciesDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Lists all notification policies.",
		Attributes: map[string]schema.Attribute{
			"notification_policies": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of notification policies.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the notification policy.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the notification policy.",
						},
					},
				},
			},
		},
	}
}

func (d *notificationPoliciesDataSource) Configure(
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

	d.client = oodlehttp.NewModelClient[*clientmodels.NotificationPolicy](
		client,
		"notification-policies",
		func() *clientmodels.NotificationPolicy { return &clientmodels.NotificationPolicy{} },
	)
}

func (d *notificationPoliciesDataSource) Read(
	ctx context.Context,
	_ datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	policies, err := d.client.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing notification policies",
			"Could not list notification policies: "+err.Error(),
		)
		return
	}

	state := notificationPoliciesDataSourceModel{
		NotificationPolicies: make([]notificationPolicyModel, 0, len(policies)),
	}
	for _, p := range policies {
		state.NotificationPolicies = append(state.NotificationPolicies, notificationPolicyModel{
			ID:   types.StringValue(p.GetID()),
			Name: types.StringValue(p.Name),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
