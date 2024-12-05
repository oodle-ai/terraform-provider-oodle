package notificationPolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
	"terraform-provider-oodle/internal/validatorutils"
)

const notificationPoliciesResource = "notification-policies"

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &notificationPolicyResource{}
	_ resource.ResourceWithConfigure   = &notificationPolicyResource{}
	_ resource.ResourceWithImportState = &notificationPolicyResource{}
)

type notificationPolicyResource struct {
	oresource.BaseResource[*clientmodels.NotificationPolicy, *notificationPolicyResourceModel]
}

func NewNotificationPolicyResource() resource.Resource {
	modelCreator := func() *clientmodels.NotificationPolicy {
		return &clientmodels.NotificationPolicy{}
	}

	return &notificationPolicyResource{
		BaseResource: oresource.NewBaseResource(
			func() *notificationPolicyResourceModel {
				return &notificationPolicyResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[*clientmodels.NotificationPolicy] {
				return oodlehttp.NewModelClient[*clientmodels.NotificationPolicy](
					oodleHttpClient,
					notificationPoliciesResource,
					modelCreator,
				)
			},
		),
	}
}

func (n *notificationPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_policy"
}

func (n *notificationPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the notification policy.",
				Validators: []validator.String{
					validatorutils.NewUUIDValidator(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the notification policy.",
			},
			"notifiers": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Notifiers by severity.",
				Attributes: map[string]schema.Attribute{
					"warn": schema.ListAttribute{
						Optional:    true,
						Description: "Notifier IDs for warning severity.",
						ElementType: types.StringType,
					},
					"critical": schema.ListAttribute{
						Optional:    true,
						Description: "Notifier IDs for critical severity.",
						ElementType: types.StringType,
					},
				},
			},
			"global": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the notification policy is a global notification policy.",
			},
			"mute_global": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to mute global notification policy.",
			},
			"mute_non_global": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to mute non-global notification policies.",
			},
		},
	}
}
