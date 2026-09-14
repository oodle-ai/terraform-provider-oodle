package gcpproject

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
	"terraform-provider-oodle/internal/validatorutils"
)

var (
	// Google requires a project ID of 6 to 30 characters: lowercase letters,
	// digits and hyphens, starting with a letter and not ending with a hyphen.
	gcpProjectPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)

	// A service account address, which is the only thing Oodle can
	// impersonate. A user address is accepted by the API and then fails every
	// scrape, so reject it at plan time.
	gcpServiceAccountPattern = regexp.MustCompile(
		`^[a-z0-9-]+@[a-z0-9-]+\.iam\.gserviceaccount\.com$`,
	)
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &gcpProjectResource{}
	_ resource.ResourceWithConfigure   = &gcpProjectResource{}
	_ resource.ResourceWithImportState = &gcpProjectResource{}
)

// gcpProjectResource uses APIBaseResource, as the Azure integration does, so
// that a read which 404s removes the resource from state instead of failing
// every later plan.
type gcpProjectResource struct {
	oresource.APIBaseResource[*clientmodels.GcpIntegration, *gcpProjectResourceModel]
}

func NewGcpProjectResource() resource.Resource {
	modelCreator := func() *clientmodels.GcpIntegration {
		return &clientmodels.GcpIntegration{}
	}
	return &gcpProjectResource{
		APIBaseResource: oresource.NewAPIBaseResource[*clientmodels.GcpIntegration, *gcpProjectResourceModel](
			func() *gcpProjectResourceModel {
				return &gcpProjectResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) oresource.ModelAPI[*clientmodels.GcpIntegration] {
				return oodlehttp.NewModelClient[*clientmodels.GcpIntegration](
					oodleHttpClient,
					clientmodels.IntegrationsResourcePath,
					modelCreator,
				)
			},
		),
	}
}

func (r *gcpProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcp_project"
}

func (r *gcpProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Oodle metric-pull integration for one GCP project. " +
			"Oodle impersonates a service account in the project to pull Cloud Monitoring metrics; " +
			"collect several projects by declaring one resource per project. " +
			"The service account must already exist and hold the viewer roles Oodle reads with, " +
			"and must grant `oodle_principal` the Service Account Token Creator role, " +
			"or metrics never arrive and the integration stays in NOT_CONNECTED.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the integration assigned by Oodle.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed: true,
				Description: "Name of the integration as shown in the Oodle UI. " +
					"Set by the server from `project`; it cannot be configured independently.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				Description: "Lifecycle status of the integration (e.g. NOT_CONNECTED, RECEIVING). " +
					"Set by the server; it becomes RECEIVING once Oodle collects metrics from the project.",
			},
			"project": schema.StringAttribute{
				Required:    true,
				Description: "ID of the GCP project to pull metrics from (the project ID, not its number or display name).",
				Validators: []validator.String{
					validatorutils.NewRegexValidator(
						gcpProjectPattern,
						"must be a GCP project ID: 6 to 30 characters of lowercase letters, digits and hyphens, starting with a letter",
					),
				},
				// The server names the row after its project when the row is
				// created and does not rename it later. A project is also a
				// different source of metrics rather than a changed setting,
				// so replace the resource instead of patching it in place.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"customer_service_account": schema.StringAttribute{
				Required: true,
				Description: "Address of the service account in the project that Oodle impersonates. " +
					"It needs the Browser, Compute Viewer, Monitoring Viewer and Cloud Asset Viewer roles on the project.",
				Validators: []validator.String{
					validatorutils.NewRegexValidator(
						gcpServiceAccountPattern,
						"must be a service account address ending in .iam.gserviceaccount.com",
					),
				},
			},
			"oodle_principal": schema.StringAttribute{
				Computed: true,
				Description: "Oodle service account that `customer_service_account` must grant the Service Account Token Creator role to. " +
					"Set by the server, which uses the same principal for every customer; reference it from the IAM grant rather than writing the address out.",
			},
		},
	}
}
