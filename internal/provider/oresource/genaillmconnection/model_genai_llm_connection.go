package genaillmconnection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type genaiLLMConnectionResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	LLMProvider         types.String `tfsdk:"llm_provider"`
	APIKey              types.String `tfsdk:"api_key"`
	BaseURL             types.String `tfsdk:"base_url"`
	DefaultModel        types.String `tfsdk:"default_model"`
	CustomModels        types.List   `tfsdk:"custom_models"`
	EnableDefaultModels types.Bool   `tfsdk:"enable_default_models"`
	IsDefault           types.Bool   `tfsdk:"is_default"`
	CustomHeaders       types.String `tfsdk:"custom_headers"`
	DefaultParams       types.String `tfsdk:"default_params"`
}

func (m *genaiLLMConnectionResourceModel) GetID() types.String {
	return m.ID
}

func (m *genaiLLMConnectionResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *genaiLLMConnectionResourceModel) FromClientModel(
	ctx context.Context,
	model *clientmodels.GenAILLMConnection,
	diagnosticsOut *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.LLMProvider = types.StringValue(model.Provider)

	// api_key and custom_headers are write-only: the API stores them
	// encrypted and omits them from every response, so the values
	// already held are kept rather than being cleared.

	m.BaseURL = optionalString(model.BaseURL, m.BaseURL)
	m.DefaultModel = optionalString(model.DefaultModel, m.DefaultModel)
	m.CustomModels = resourceutils.SliceToStringList(
		ctx, model.CustomModels, m.CustomModels, diagnosticsOut,
	)

	if model.EnableDefaultModels != nil {
		m.EnableDefaultModels = types.BoolValue(*model.EnableDefaultModels)
	} else {
		m.EnableDefaultModels = types.BoolValue(false)
	}

	if model.IsDefault != nil {
		m.IsDefault = types.BoolValue(*model.IsDefault)
	} else {
		m.IsDefault = types.BoolValue(false)
	}

	m.DefaultParams = resourceutils.RawToJSONString(
		model.DefaultParams, m.DefaultParams,
	)
}

func (m *genaiLLMConnectionResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.GenAILLMConnection,
) error {
	model.ID = m.ID.ValueString()
	model.Name = m.Name.ValueString()
	model.Provider = m.LLMProvider.ValueString()
	model.APIKey = m.APIKey.ValueString()
	model.BaseURL = m.BaseURL.ValueString()
	model.DefaultModel = m.DefaultModel.ValueString()

	customModels, err := resourceutils.StringListToSlice(ctx, m.CustomModels)
	if err != nil {
		return err
	}
	model.CustomModels = customModels

	// The update endpoint replaces these only when they are present,
	// so they are always sent to make the applied state match the
	// configuration.
	enableDefaultModels := m.EnableDefaultModels.ValueBool()
	model.EnableDefaultModels = &enableDefaultModels

	isDefault := m.IsDefault.ValueBool()
	model.IsDefault = &isDefault

	customHeaders, err := resourceutils.JSONStringToRaw(
		m.CustomHeaders, "custom_headers",
	)
	if err != nil {
		return err
	}
	model.CustomHeaders = customHeaders

	defaultParams, err := resourceutils.JSONStringToRaw(
		m.DefaultParams, "default_params",
	)
	if err != nil {
		return err
	}
	model.DefaultParams = defaultParams

	return nil
}

// optionalString keeps an unset optional attribute null instead of
// turning it into "" when the API omits the field.
func optionalString(value string, prior types.String) types.String {
	if value == "" && (prior.IsNull() || prior.IsUnknown()) {
		return types.StringNull()
	}

	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}
