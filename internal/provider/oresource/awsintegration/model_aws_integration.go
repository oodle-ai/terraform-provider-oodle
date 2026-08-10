package awsintegration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type awsIntegrationResourceModel struct {
	ID                      types.String                  `tfsdk:"id"`
	Name                    types.String                  `tfsdk:"name"`
	Status                  types.String                  `tfsdk:"status"`
	AccountID               types.String                  `tfsdk:"account_id"`
	RoleArn                 types.String                  `tfsdk:"role_arn"`
	ExternalID              types.String                  `tfsdk:"external_id"`
	Regions                 []types.String                `tfsdk:"regions"`
	LaunchCFStackRegion     types.String                  `tfsdk:"launch_cf_stack_region"`
	LaunchCFStackURL        types.String                  `tfsdk:"launch_cf_stack_url"`
	ResourceTypesSearchTags []resourceTypeSearchTagsModel `tfsdk:"resource_types_search_tags"`
}

type resourceTypeSearchTagsModel struct {
	ResourceTypes []types.String `tfsdk:"resource_types"`
	// SearchTags is a types.List (not a native []searchTagModel) because the
	// attribute is Optional+Computed: when the block is omitted the framework
	// plans it as an *unknown* value, which a native Go slice cannot decode
	// ("Received unknown value, however the target type cannot handle unknown
	// values"). A types.List can carry null/unknown/known states, so it round-
	// trips through Plan.Get without error. Elements are searchTagModel objects.
	SearchTags types.List `tfsdk:"search_tags"`
}

type searchTagModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

// searchTagAttrTypes / searchTagObjectType describe the element shape of the
// search_tags list so it can be (de)serialized as a types.List.
var searchTagAttrTypes = map[string]attr.Type{
	"key":   types.StringType,
	"value": types.StringType,
}

var searchTagObjectType = types.ObjectType{AttrTypes: searchTagAttrTypes}

var _ resourceutils.ResourceModel[*clientmodels.AwsIntegration] = (*awsIntegrationResourceModel)(nil)

func (m *awsIntegrationResourceModel) GetID() types.String {
	return m.ID
}

func (m *awsIntegrationResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *awsIntegrationResourceModel) FromClientModel(
	_ context.Context,
	model *clientmodels.AwsIntegration,
	diags *diag.Diagnostics,
) {
	*m = awsIntegrationResourceModel{}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.Status = types.StringValue(model.Status)

	cw := model.TypeSpecificData.CloudWatchMetricPullIntegration
	m.AccountID = types.StringValue(cw.AccountID)
	m.RoleArn = types.StringValue(cw.RoleArn)
	m.ExternalID = types.StringValue(cw.ExternalID)
	m.LaunchCFStackRegion = types.StringValue(cw.LaunchCFStackRegion)
	m.LaunchCFStackURL = types.StringValue(cw.LaunchCFStackURL)

	// Required schema fields: always allocate so an empty server response
	// round-trips as []string{} instead of nil and avoids perpetual plan
	// diff on `terraform import` of an integration that has no regions yet.
	m.Regions = make([]types.String, len(cw.Regions))
	for i, r := range cw.Regions {
		m.Regions[i] = types.StringValue(r)
	}

	m.ResourceTypesSearchTags = make([]resourceTypeSearchTagsModel, len(cw.ResourceTypesSearchTagsList))
	for i, entry := range cw.ResourceTypesSearchTagsList {
		row := resourceTypeSearchTagsModel{
			ResourceTypes: make([]types.String, len(entry.ResourceTypes)),
		}
		for j, t := range entry.ResourceTypes {
			row.ResourceTypes[j] = types.StringValue(t)
		}
		// Always build a known, non-null list (even when empty) so a server
		// response with no tags round-trips as [] instead of null. Without this,
		// a config that sets `search_tags = []` fails Create with "inconsistent
		// result after apply" because the applied state (null) would differ from
		// the plan ([]). Mirrors the Regions/ResourceTypes handling above; paired
		// with the Optional+Computed schema so an omitted block is tolerated too.
		tagValues := make([]attr.Value, len(entry.SearchTags))
		for j, tag := range entry.SearchTags {
			obj, d := types.ObjectValue(searchTagAttrTypes, map[string]attr.Value{
				"key":   types.StringValue(tag.Key),
				"value": types.StringValue(tag.Value),
			})
			diags.Append(d...)
			tagValues[j] = obj
		}
		list, d := types.ListValue(searchTagObjectType, tagValues)
		diags.Append(d...)
		row.SearchTags = list
		m.ResourceTypesSearchTags[i] = row
	}
}

func (m *awsIntegrationResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.AwsIntegration,
) error {
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		model.ID = m.ID.ValueString()
	}

	// AWS integrations are always written with the CLOUDWATCH_METRIC_PULL
	// discriminator; the user does not set this on the resource.
	model.Type = clientmodels.AwsIntegrationType

	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		model.Name = m.Name.ValueString()
	}
	// status and launch_cf_stack_url are Computed-only — server controls
	// them and we never echo them back in Create/Update bodies.

	cw := clientmodels.CloudWatchMetricPullIntegration{
		AccountID:  m.AccountID.ValueString(),
		RoleArn:    m.RoleArn.ValueString(),
		ExternalID: m.ExternalID.ValueString(),
	}
	if !m.LaunchCFStackRegion.IsNull() && !m.LaunchCFStackRegion.IsUnknown() {
		cw.LaunchCFStackRegion = m.LaunchCFStackRegion.ValueString()
	}

	if len(m.Regions) > 0 {
		cw.Regions = make([]string, len(m.Regions))
		for i, r := range m.Regions {
			cw.Regions[i] = r.ValueString()
		}
	}

	if len(m.ResourceTypesSearchTags) > 0 {
		cw.ResourceTypesSearchTagsList = make([]clientmodels.CloudWatchResourceTypeSearchTags, len(m.ResourceTypesSearchTags))
		for i, row := range m.ResourceTypesSearchTags {
			entry := clientmodels.CloudWatchResourceTypeSearchTags{}
			if len(row.ResourceTypes) > 0 {
				entry.ResourceTypes = make([]string, len(row.ResourceTypes))
				for j, t := range row.ResourceTypes {
					entry.ResourceTypes[j] = t.ValueString()
				}
			}
			// The list is unknown when an Optional+Computed block is omitted on
			// Create; treat that (and null) as "no tags" and let the server
			// normalize. Only decode elements when the value is known.
			if !row.SearchTags.IsNull() && !row.SearchTags.IsUnknown() {
				var tags []searchTagModel
				if d := row.SearchTags.ElementsAs(ctx, &tags, false); d.HasError() {
					return fmt.Errorf("failed to read search_tags for resource_types_search_tags[%d]: %v", i, d.Errors())
				}
				if len(tags) > 0 {
					entry.SearchTags = make([]clientmodels.CloudWatchSearchTag, len(tags))
					for j, tag := range tags {
						entry.SearchTags[j] = clientmodels.CloudWatchSearchTag{
							Key:   tag.Key.ValueString(),
							Value: tag.Value.ValueString(),
						}
					}
				}
			}
			cw.ResourceTypesSearchTagsList[i] = entry
		}
	}

	model.TypeSpecificData.CloudWatchMetricPullIntegration = cw
	return nil
}
