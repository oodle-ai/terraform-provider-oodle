package genaicodelibrary

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

type genaiCodeLibraryResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	SourceCode  types.String `tfsdk:"source_code"`
	Version     types.Int64  `tfsdk:"version"`
}

func (m *genaiCodeLibraryResourceModel) GetID() types.String {
	return m.ID
}

func (m *genaiCodeLibraryResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *genaiCodeLibraryResourceModel) FromClientModel(
	_ context.Context,
	model *clientmodels.GenAICodeLibrary,
	_ *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.SourceCode = types.StringValue(derefString(model.SourceCode))
	m.Version = types.Int64Value(model.Version)

	// The API always answers with a description, "" when there is
	// none. An unset attribute stays null, so that it does not show
	// a diff against "".
	description := derefString(model.Description)
	if description == "" && m.Description.ValueString() == "" &&
		(m.Description.IsNull() || m.Description.IsUnknown()) {
		m.Description = types.StringNull()
	} else {
		m.Description = types.StringValue(description)
	}
}

func (m *genaiCodeLibraryResourceModel) ToClientModel(
	_ context.Context,
	model *clientmodels.GenAICodeLibrary,
) error {
	model.ID = m.ID.ValueString()
	model.Name = m.Name.ValueString()

	sourceCode := m.SourceCode.ValueString()
	model.SourceCode = &sourceCode

	// The update endpoint leaves an omitted description alone, so ""
	// is sent to clear one that was removed from the configuration.
	description := m.Description.ValueString()
	model.Description = &description

	return nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
