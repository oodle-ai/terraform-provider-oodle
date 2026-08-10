package awsintegration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestAwsIntegrationModelRoundTrip(t *testing.T) {
	ctx := context.Background()
	// Status and LaunchCFStackURL are Computed-only on the resource and
	// ToClientModel intentionally never writes them, so they're omitted
	// from this round-trip fixture; see TestAwsIntegrationModelFromServer
	// for coverage that those fields propagate from a server response
	// into the Terraform model.
	clientModel := &clientmodels.AwsIntegration{
		ID:   "int-abc123",
		Type: clientmodels.AwsIntegrationType,
		Name: "prod-aws",
	}
	clientModel.TypeSpecificData.CloudWatchMetricPullIntegration = clientmodels.CloudWatchMetricPullIntegration{
		AccountID:           "123456789012",
		LaunchCFStackRegion: "us-west-2",
		RoleArn:             "arn:aws:iam::123456789012:role/OodleIntegrationRole",
		ExternalID:          "ext-uuid-7-value",
		Regions:             []string{"us-west-2", "us-east-1"},
		ResourceTypesSearchTagsList: []clientmodels.CloudWatchResourceTypeSearchTags{
			{
				ResourceTypes: []string{"AWS/EC2", "AWS/RDS"},
				SearchTags: []clientmodels.CloudWatchSearchTag{
					{Key: "Environment", Value: "prod"},
					{Key: "Team", Value: "platform-.*"},
				},
			},
			{
				ResourceTypes: []string{"AWS/Lambda"},
			},
		},
	}

	resourceModel := &awsIntegrationResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.AwsIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}

// TestAwsIntegrationModelFromServer verifies that Computed-only fields
// returned by the server (Status, LaunchCFStackURL) land in the
// Terraform model so they're visible in state, even though ToClientModel
// does not echo them back.
func TestAwsIntegrationModelFromServer(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.AwsIntegration{
		ID:     "int-server-set",
		Type:   clientmodels.AwsIntegrationType,
		Name:   "prod-aws",
		Status: "RECEIVING",
	}
	clientModel.TypeSpecificData.CloudWatchMetricPullIntegration = clientmodels.CloudWatchMetricPullIntegration{
		AccountID:           "123456789012",
		LaunchCFStackURL:    "https://console.aws.amazon.com/cloudformation/home?region=us-west-2#/stacks/quickcreate",
		LaunchCFStackRegion: "us-west-2",
		RoleArn:             "arn:aws:iam::123456789012:role/OodleIntegrationRole",
		ExternalID:          "ext-uuid-7-value",
		Regions:             []string{"us-west-2"},
		ResourceTypesSearchTagsList: []clientmodels.CloudWatchResourceTypeSearchTags{
			{ResourceTypes: []string{"AWS/EC2"}},
		},
	}

	resourceModel := &awsIntegrationResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.Equal(t, "RECEIVING", resourceModel.Status.ValueString())
	assert.Equal(
		t,
		"https://console.aws.amazon.com/cloudformation/home?region=us-west-2#/stacks/quickcreate",
		resourceModel.LaunchCFStackURL.ValueString(),
	)

	// Confirm ToClientModel does NOT echo these back to the server.
	echoed := &clientmodels.AwsIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, echoed))
	assert.Equal(t, "", echoed.Status)
	assert.Equal(t, "", echoed.TypeSpecificData.CloudWatchMetricPullIntegration.LaunchCFStackURL)
}

// TestAwsIntegrationModelEmptySearchTagsNonNil guards the fix for the
// "Provider produced inconsistent result after apply" error: when the server
// returns an entry with no search tags, FromClientModel must produce a non-nil
// empty slice so a config that sets `search_tags = []` round-trips as [] rather
// than collapsing to null in state.
func TestAwsIntegrationModelEmptySearchTagsNonNil(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.AwsIntegration{
		ID:   "int-empty-tags",
		Type: clientmodels.AwsIntegrationType,
	}
	clientModel.TypeSpecificData.CloudWatchMetricPullIntegration = clientmodels.CloudWatchMetricPullIntegration{
		AccountID:  "123456789012",
		RoleArn:    "arn:aws:iam::123456789012:role/OodleIntegrationRole",
		ExternalID: "shared-ext-id",
		Regions:    []string{"eu-central-1"},
		ResourceTypesSearchTagsList: []clientmodels.CloudWatchResourceTypeSearchTags{
			{ResourceTypes: []string{"AWS/DynamoDB", "AWS/Lambda"}},
		},
	}

	resourceModel := &awsIntegrationResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.Equal(t, 1, len(resourceModel.ResourceTypesSearchTags))
	// Known and non-null (so it renders as [] not null) but empty.
	assert.False(t, resourceModel.ResourceTypesSearchTags[0].SearchTags.IsNull())
	assert.False(t, resourceModel.ResourceTypesSearchTags[0].SearchTags.IsUnknown())
	assert.Equal(t, 0, len(resourceModel.ResourceTypesSearchTags[0].SearchTags.Elements()))
}

// TestAwsIntegrationModelUnknownSearchTags guards the fix for the
// "Received unknown value, however the target type cannot handle unknown
// values" error. search_tags is Optional+Computed, so when the block is
// omitted in config the framework plans it as an *unknown* list. That unknown
// must survive Plan.Get into the model (hence SearchTags is a types.List, not a
// native slice) and ToClientModel must treat it as "no tags" rather than
// panicking or emitting garbage.
func TestAwsIntegrationModelUnknownSearchTags(t *testing.T) {
	ctx := context.Background()
	resourceModel := &awsIntegrationResourceModel{
		AccountID:  types.StringValue("123456789012"),
		RoleArn:    types.StringValue("arn:aws:iam::123456789012:role/OodleIntegrationRole"),
		ExternalID: types.StringValue("shared-ext-id"),
		Regions:    []types.String{types.StringValue("ap-south-1")},
		ResourceTypesSearchTags: []resourceTypeSearchTagsModel{
			{
				ResourceTypes: []types.String{types.StringValue("AWS/EC2")},
				// Mirrors an omitted Optional+Computed block during Create.
				SearchTags: types.ListUnknown(searchTagObjectType),
			},
		},
	}

	clientModel := &clientmodels.AwsIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, clientModel))

	cw := clientModel.TypeSpecificData.CloudWatchMetricPullIntegration
	assert.Equal(t, 1, len(cw.ResourceTypesSearchTagsList))
	assert.DeepEqual(t, []string{"AWS/EC2"}, cw.ResourceTypesSearchTagsList[0].ResourceTypes)
	// Unknown search_tags is sent as no tags; the server normalizes.
	assert.Nil(t, cw.ResourceTypesSearchTagsList[0].SearchTags)
}

func TestAwsIntegrationModelMinimal(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.AwsIntegration{
		ID:   "int-min",
		Type: clientmodels.AwsIntegrationType,
	}
	clientModel.TypeSpecificData.CloudWatchMetricPullIntegration = clientmodels.CloudWatchMetricPullIntegration{
		AccountID:  "210987654321",
		RoleArn:    "arn:aws:iam::210987654321:role/OodleIntegrationRole",
		ExternalID: "shared-ext-id",
		Regions:    []string{"us-west-2"},
		ResourceTypesSearchTagsList: []clientmodels.CloudWatchResourceTypeSearchTags{
			{ResourceTypes: []string{"AWS/EC2"}},
		},
	}

	resourceModel := &awsIntegrationResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.AwsIntegration{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}
