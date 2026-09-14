package azureintegration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
	"terraform-provider-oodle/internal/validatorutils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &azureIntegrationResource{}
	_ resource.ResourceWithConfigure   = &azureIntegrationResource{}
	_ resource.ResourceWithImportState = &azureIntegrationResource{}
)

// azureIntegrationResource uses APIBaseResource rather than the BaseResource
// used by the AWS integration, for two reasons specific to this API. The plan
// is the receiver when converting a response back into state, so the
// write-only client_secret keeps its configured value instead of being
// overwritten by the mask the backend returns. And a read that 404s removes
// the resource from state rather than failing every subsequent plan.
type azureIntegrationResource struct {
	oresource.APIBaseResource[*clientmodels.AzureIntegration, *azureIntegrationResourceModel]
}

func NewAzureIntegrationResource() resource.Resource {
	modelCreator := func() *clientmodels.AzureIntegration {
		return &clientmodels.AzureIntegration{}
	}
	return &azureIntegrationResource{
		APIBaseResource: oresource.NewAPIBaseResource[*clientmodels.AzureIntegration, *azureIntegrationResourceModel](
			func() *azureIntegrationResourceModel {
				return &azureIntegrationResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) oresource.ModelAPI[*clientmodels.AzureIntegration] {
				return oodlehttp.NewModelClient[*clientmodels.AzureIntegration](
					oodleHttpClient,
					clientmodels.IntegrationsResourcePath,
					modelCreator,
				)
			},
		),
	}
}

func (r *azureIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_integration"
}

func (r *azureIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Oodle Azure Monitor metric-pull integration. " +
			"Oodle authenticates as an Azure app registration to pull platform metrics for one subscription; " +
			"manage several subscriptions by declaring one resource per subscription. " +
			"The app registration must already exist and hold the Monitoring Reader role on the subscription (or on the resource groups being collected). " +
			"Oodle verifies the credentials against Azure when the resource is created or updated, so an apply fails fast if the role is missing.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the Azure integration assigned by Oodle.",
			},
			"name": schema.StringAttribute{
				Computed: true,
				Description: "Name of the integration as shown in the Oodle UI. " +
					"Set by the server from `subscription_name`; it cannot be configured independently.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				Description: "Lifecycle status of the integration (e.g. NOT_CONNECTED, RECEIVING). " +
					"Set by the server; transitions to RECEIVING once Oodle begins collecting metrics, which can take up to 15 minutes after creation.",
			},
			"subscription_name": schema.StringAttribute{
				Required: true,
				Description: "Human-readable label for the subscription, shown in the Oodle UI. " +
					"The integration's `name` is derived from this value.",
			},
			"tenant_id": schema.StringAttribute{
				Required:    true,
				Description: "Azure Entra tenant ID that the app registration belongs to.",
			},
			"subscription_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the Azure subscription to pull metrics from.",
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "Application (client) ID of the Azure app registration Oodle authenticates as.",
			},
			"client_secret": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				Description: "Client secret for the app registration. " +
					"Write-only: the API stores this encrypted and returns a mask instead of the value, so Terraform cannot detect a secret rotated outside of Terraform. " +
					"After `terraform import` this attribute is absent from state and the next plan will show a diff until it is set in the configuration.",
			},
			"service_filters": schema.ListNestedAttribute{
				Optional: true,
				Description: "Azure services to collect metrics for, grouped so each group can carry its own tag filters. " +
					"Omit to collect Oodle's default set of services.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"service_ids": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
							Description: "Oodle service identifiers to collect, e.g. [\"compute_virtualmachines\", \"storage_storageaccounts\"]. Unknown IDs are rejected by the server.",
						},
						"tags": schema.MapAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Optional Azure resource tag filters. Include-only: a resource must carry every tag listed here to be collected by this group.",
						},
					},
				},
			},
			"resource_groups": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Restrict collection to these Azure resource groups. Omit to collect across the whole subscription.",
			},
			"resource_name_regex": schema.StringAttribute{
				Optional: true,
				Description: "Restrict collection to resources whose name matches this regular expression. " +
					"Omit to collect every resource.",
				Validators: []validator.String{
					validatorutils.NewValidRegexValidator(),
				},
			},
		},
	}
}
