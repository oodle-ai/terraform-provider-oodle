package awsintegration

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/provider/oresource"
	"terraform-provider-oodle/internal/validatorutils"
)

var (
	awsAccountIDPattern = regexp.MustCompile(`^\d{12}$`)
	awsRoleArnPattern   = regexp.MustCompile(`^arn:aws:iam::\d{12}:role/.+$`)
	// Oodle can only assume roles whose name begins with "OodleIntegration"
	// (the name used by the Oodle CloudFormation setup). A role with any other
	// name cannot be assumed, which otherwise only surfaces at apply time as the
	// opaque "Failed to assume IAM role" backend error; validate the name up
	// front so it is caught at plan time instead.
	awsRoleNamePattern = regexp.MustCompile(`^arn:aws:iam::\d{12}:role/OodleIntegration`)
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &awsIntegrationResource{}
	_ resource.ResourceWithConfigure   = &awsIntegrationResource{}
	_ resource.ResourceWithImportState = &awsIntegrationResource{}
)

// integrationsResourcePath is the URL segment under
// /v1/api/instance/{instance}/ that the backend uses for integration CRUD.
// The AWS-specific shape is carried inside the body via the `type` and
// `typeSpecificData.cloudWatchMetricPullIntegration` fields.
const integrationsResourcePath = "integrations"

type awsIntegrationResource struct {
	oresource.BaseResource[*clientmodels.AwsIntegration, *awsIntegrationResourceModel]
}

func NewAwsIntegrationResource() resource.Resource {
	modelCreator := func() *clientmodels.AwsIntegration {
		return &clientmodels.AwsIntegration{}
	}
	return &awsIntegrationResource{
		BaseResource: oresource.NewBaseResource[*clientmodels.AwsIntegration, *awsIntegrationResourceModel](
			func() *awsIntegrationResourceModel {
				return &awsIntegrationResourceModel{}
			},
			modelCreator,
			func(oodleHttpClient *oodlehttp.OodleApiClient) *oodlehttp.ModelClient[*clientmodels.AwsIntegration] {
				return oodlehttp.NewModelClient[*clientmodels.AwsIntegration](
					oodleHttpClient,
					integrationsResourcePath,
					modelCreator,
				)
			},
		),
	}
}

func (r *awsIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_integration"
}

func (r *awsIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Oodle AWS CloudWatch metric-pull integration. " +
			"Oodle assumes the given IAM role to pull CloudWatch metrics for the configured regions and resource types. " +
			"The IAM role must already exist with a trust policy that allows Oodle's AWS account to assume it under the given external ID; " +
			"see the `oodle_aws_integration` module under examples/modules for a one-shot deploy that creates the role via CloudFormation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the AWS integration assigned by Oodle.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Human-readable name for the integration. If omitted, the server assigns one.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				Description: "Lifecycle status of the integration (e.g. INACTIVE, RECEIVING). " +
					"Set by the server; transitions automatically once Oodle successfully assumes the IAM role and begins receiving metrics.",
			},
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "12-digit AWS account ID to pull metrics from.",
				Validators: []validator.String{
					validatorutils.NewRegexValidator(awsAccountIDPattern, "must be a 12-digit AWS account ID"),
				},
			},
			"role_arn": schema.StringAttribute{
				Required: true,
				Description: "ARN of the IAM role Oodle assumes in the target account. " +
					"The role's trust policy must allow Oodle's AWS account (052799302239) to assume it under the configured external_id.",
				Validators: []validator.String{
					validatorutils.NewRegexValidator(awsRoleArnPattern, "must be an IAM role ARN (arn:aws:iam::<account-id>:role/<role-name>)"),
					validatorutils.NewRegexValidator(awsRoleNamePattern, `must name an IAM role beginning with "OodleIntegration"; Oodle can only assume roles with this name prefix`),
				},
			},
			"external_id": schema.StringAttribute{
				Required: true,
				Description: "External ID used as a condition in the IAM role's trust policy. " +
					"Share the same value across all AWS integrations in the workspace so a single CloudFormation deploy covers every account; " +
					"a `random_uuid` resource is the recommended way to generate it.",
			},
			"regions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "AWS regions to pull metrics from (e.g. [\"us-west-2\", \"us-east-1\"]).",
			},
			"launch_cf_stack_region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Region used when constructing the saved CloudFormation launch URL displayed in the Oodle UI. " +
					"Does not affect which regions metrics are pulled from. Defaults to us-west-2 if omitted.",
			},
			"launch_cf_stack_url": schema.StringAttribute{
				Computed:    true,
				Description: "Server-rendered CloudFormation launch URL displayed in the Oodle UI for users that have not yet provisioned the IAM role.",
			},
			"resource_types_search_tags": schema.ListNestedAttribute{
				Required:    true,
				Description: "Resource type groups Oodle should discover and pull metrics for. Each entry pairs a list of CloudWatch namespaces with optional tag-based filters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource_types": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
							Description: "CloudWatch namespaces (e.g. [\"AWS/EC2\", \"AWS/RDS\"]).",
						},
						"search_tags": schema.ListNestedAttribute{
							Optional: true,
							// Computed so the server's null->empty normalization
							// does not produce an inconsistent-result error when
							// the block is omitted; an explicit [] round-trips as
							// [] via FromClientModel.
							Computed:    true,
							Description: "Optional tag filters; all listed tags must match for a resource to be included. Values may be regular expressions.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Required:    true,
										Description: "Tag key to match.",
									},
									"value": schema.StringAttribute{
										Required:    true,
										Description: "Tag value to match. May be a regular expression.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
