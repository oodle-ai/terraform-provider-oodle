package genaievaltemplate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type genaiEvalTemplateResourceModel struct {
	ID                 types.String       `tfsdk:"id"`
	Name               types.String       `tfsdk:"name"`
	Type               types.String       `tfsdk:"type"`
	Prompt             types.String       `tfsdk:"prompt"`
	Vars               types.List         `tfsdk:"vars"`
	OutputSchema       resourceutils.JSON `tfsdk:"output_schema"`
	ModelParams        resourceutils.JSON `tfsdk:"model_params"`
	SourceCode         types.String       `tfsdk:"source_code"`
	SourceCodeLanguage types.String       `tfsdk:"source_code_language"`
	Version            types.Int64        `tfsdk:"version"`
}

func (m *genaiEvalTemplateResourceModel) GetID() types.String {
	return m.ID
}

func (m *genaiEvalTemplateResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *genaiEvalTemplateResourceModel) FromClientModel(
	ctx context.Context,
	model *clientmodels.GenAIEvalTemplate,
	diagnosticsOut *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.Type = types.StringValue(model.Type)
	m.Prompt = optionalString(model.Prompt)
	m.SourceCode = optionalString(model.SourceCode)
	m.SourceCodeLanguage = optionalString(model.SourceCodeLanguage)
	m.Version = types.Int64Value(model.Version)

	m.Vars = resourceutils.SliceToStringList(
		ctx, model.Vars, m.Vars, diagnosticsOut,
	)
	m.OutputSchema = resourceutils.RawToJSON(model.OutputSchema)
	m.ModelParams = resourceutils.RawToJSON(model.ModelParams)
}

func (m *genaiEvalTemplateResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.GenAIEvalTemplate,
) error {
	model.ID = m.ID.ValueString()
	model.Name = m.Name.ValueString()
	model.Type = m.Type.ValueString()
	model.Prompt = m.Prompt.ValueString()
	model.SourceCode = m.SourceCode.ValueString()
	model.SourceCodeLanguage = m.SourceCodeLanguage.ValueString()

	vars, err := resourceutils.StringListToSlice(ctx, m.Vars)
	if err != nil {
		return err
	}
	model.Vars = vars

	outputSchema, err := resourceutils.JSONToRaw(
		m.OutputSchema, "output_schema",
	)
	if err != nil {
		return err
	}
	model.OutputSchema = outputSchema

	modelParams, err := resourceutils.JSONToRaw(
		m.ModelParams, "model_params",
	)
	if err != nil {
		return err
	}
	model.ModelParams = modelParams

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
