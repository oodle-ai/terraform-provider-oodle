package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/provider/oresource/monitor"
	"terraform-provider-oodle/internal/provider/oresource/notifier"
)

const (
	deploymentUrlField = "deployment_url"
	instanceField      = "instance"
	apiKeyField        = "api_key"
)

// oodleProviderModel maps provider schema data to a Go type.
type oodleProviderModel struct {
	DeploymentUrl types.String `tfsdk:"deployment_url"`
	Instance      types.String `tfsdk:"instance"`
	APIKey        types.String `tfsdk:"api_key"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &oodleProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &oodleProvider{
			version: version,
		}
	}
}

// oodleProvider is the provider implementation.
type oodleProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *oodleProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oodle"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *oodleProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			deploymentUrlField: schema.StringAttribute{
				Optional: true,
			},
			instanceField: schema.StringAttribute{
				Optional: true,
			},
			apiKeyField: schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *oodleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring HashiCups client")
	// Retrieve provider data from configuration
	var config oodleProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.DeploymentUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(deploymentUrlField),
			"Unknown Oodle Deployment",
			"The provider cannot create the Oodle API client as there is an unknown configuration value for the Oodle Deployment. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the OODLE_DEPLOYMENT environment variable.",
		)
	}

	if config.Instance.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(instanceField),
			"Unknown Oodle instance",
			"The provider cannot create the Oodle API client as there is an unknown configuration value for the Oodle instance. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the OODLE_INSTANCE environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(apiKeyField),
			"Unknown Oodle API Key",
			"The provider cannot create the Oodle API client as there is an unknown configuration value for the Oodle API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the OODLE_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	deployment := os.Getenv("OODLE_DEPLOYMENT")
	instance := os.Getenv("OODLE_INSTANCE")
	apiKey := os.Getenv("OODLE_API_KEY")

	if !config.DeploymentUrl.IsNull() {
		deployment = config.DeploymentUrl.ValueString()
	}

	if !config.Instance.IsNull() {
		instance = config.Instance.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	ctx = tflog.SetField(ctx, deploymentUrlField, deployment)
	ctx = tflog.SetField(ctx, instanceField, instance)
	ctx = tflog.SetField(ctx, apiKeyField, apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, apiKeyField)

	tflog.Debug(ctx, "Creating Oodle client")

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if deployment == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root(deploymentUrlField),
			"Missing Oodle Deployment",
			"The provider cannot create the Oodle API client as there is a missing or empty value for the Deployment. "+
				"Set the deployment value in the configuration or use the OODLE_DEPLOYMENT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if instance == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root(instanceField),
			"Missing Oodle Instance",
			"The provider cannot create the Oodle API client as there is a missing or empty value for the Oodle Instance. "+
				"Set the instance value in the configuration or use the OODLE_INSTANCE environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root(apiKeyField),
			"Missing Oodle API Key",
			"The provider cannot create the Oodle API client as there is a missing or empty value for the Oodle API key. "+
				"Set the api key value in the configuration or use the OODLE_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Oodle client using the configuration values
	client, err := oodlehttp.NewInstanceClient(deployment, instance, apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Oodle API OodleApiClient",
			"An unexpected error occurred when creating the Oodle API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Oodle OodleApiClient Error: "+err.Error(),
		)
		return
	}

	// Make the Oodle client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
	tflog.Info(ctx, "Configured Oodle client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *oodleProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *oodleProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		monitor.NewMonitorResource,
		notifier.NewNotifierResource,
		//NewNotiicationPolicyResource,
	}
}
