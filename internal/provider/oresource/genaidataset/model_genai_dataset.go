package genaidataset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type genaiDatasetResourceModel struct {
	// ID holds the dataset name: the read and delete endpoints are
	// keyed by name, and the uuid is exposed as DatasetID.
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Metadata    types.String `tfsdk:"metadata"`
	DatasetID   types.String `tfsdk:"dataset_id"`
}

func (m *genaiDatasetResourceModel) GetID() types.String {
	return m.ID
}

func (m *genaiDatasetResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *genaiDatasetResourceModel) FromClientModel(
	_ context.Context,
	model *clientmodels.GenAIDataset,
	_ *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.Name)
	m.Name = types.StringValue(model.Name)
	m.DatasetID = types.StringValue(model.ID)

	if model.Description == "" {
		m.Description = types.StringNull()
	} else {
		m.Description = types.StringValue(model.Description)
	}

	m.Metadata = resourceutils.RawToJSONString(model.Metadata, m.Metadata)
}

func (m *genaiDatasetResourceModel) ToClientModel(
	_ context.Context,
	model *clientmodels.GenAIDataset,
) error {
	model.Name = m.Name.ValueString()
	model.Description = m.Description.ValueString()

	metadata, err := resourceutils.JSONStringToRaw(m.Metadata, "metadata")
	if err != nil {
		return err
	}
	model.Metadata = metadata

	return nil
}
