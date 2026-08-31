package genaidatasetitem

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type genaiDatasetItemResourceModel struct {
	ID                  types.String         `tfsdk:"id"`
	DatasetName         types.String         `tfsdk:"dataset_name"`
	Input               jsontypes.Normalized `tfsdk:"input"`
	ExpectedOutput      jsontypes.Normalized `tfsdk:"expected_output"`
	Metadata            jsontypes.Normalized `tfsdk:"metadata"`
	Status              types.String         `tfsdk:"status"`
	SourceTraceID       types.String         `tfsdk:"source_trace_id"`
	SourceObservationID types.String         `tfsdk:"source_observation_id"`
	DatasetID           types.String         `tfsdk:"dataset_id"`
}

func (m *genaiDatasetItemResourceModel) GetID() types.String {
	return m.ID
}

func (m *genaiDatasetItemResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *genaiDatasetItemResourceModel) FromClientModel(
	_ context.Context,
	model *clientmodels.GenAIDatasetItem,
	_ *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.DatasetID = types.StringValue(model.DatasetID)
	m.Status = types.StringValue(model.Status)

	// The API answers with the dataset uuid rather than the name it
	// was created with, so the configured name is kept.
	if model.DatasetName != "" {
		m.DatasetName = types.StringValue(model.DatasetName)
	}

	m.SourceTraceID = optionalString(model.SourceTraceID)
	m.SourceObservationID = optionalString(model.SourceObservationID)

	m.Input = resourceutils.RawToNormalized(model.Input)
	m.ExpectedOutput = resourceutils.RawToNormalized(model.ExpectedOutput)
	m.Metadata = resourceutils.RawToNormalized(model.Metadata)
}

func (m *genaiDatasetItemResourceModel) ToClientModel(
	_ context.Context,
	model *clientmodels.GenAIDatasetItem,
) error {
	model.ID = m.ID.ValueString()
	model.DatasetName = m.DatasetName.ValueString()
	model.Status = m.Status.ValueString()
	model.SourceTraceID = m.SourceTraceID.ValueString()
	model.SourceObservationID = m.SourceObservationID.ValueString()

	input, err := resourceutils.NormalizedToRaw(m.Input, "input")
	if err != nil {
		return err
	}
	model.Input = input

	expectedOutput, err := resourceutils.NormalizedToRaw(
		m.ExpectedOutput, "expected_output",
	)
	if err != nil {
		return err
	}
	model.ExpectedOutput = expectedOutput

	metadata, err := resourceutils.NormalizedToRaw(m.Metadata, "metadata")
	if err != nil {
		return err
	}
	model.Metadata = metadata

	return nil
}

// optionalString keeps an unset optional attribute null rather than
// turning it into "" when the API omits the field.
func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}
