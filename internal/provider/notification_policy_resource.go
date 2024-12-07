package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-oodle/internal/validatorutils"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &notificationPolicyResource{}
	//_ resource.ResourceWithConfigure   = &notificationPolicyResource{}
	//_ resource.ResourceWithImportState = &notificationPolicyResource{}
)

type notificationPolicyResource struct {
	resource2.baseResource[*clientmodels.NotificationPolicy, *notificationPolicyResourceModel]
}

func NewNotiicationPolicyResource() resource.Resource {
	return &notificationPolicyResource{}
}

type notificationPolicyResourceModel struct {
	ID types.String `tfsdk:"id"`
}

var _ resource.ResourceModel = &notificationPolicyResourceModel{}

func (n *notificationPolicyResourceModel) GetID() types.String {
	return n.ID
}

func (n *notificationPolicyResourceModel) SetID(id types.String) {
	n.ID = id
}

func (n *notificationPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notificationPolicy"
}

func (n *notificationPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the monitor.",
				Validators: []validator.String{
					validatorutils.NewUUIDValidator(),
				},
			},
		},
	}
}

func (n *notificationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (n *notificationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (n *notificationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (n *notificationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}
