package awsintegration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
