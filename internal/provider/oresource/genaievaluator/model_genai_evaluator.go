package genaievaluator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type genaiEvaluatorResourceModel struct {
	ID                    types.String  `tfsdk:"id"`
	Name                  types.String  `tfsdk:"name"`
	EvalTemplateID        types.String  `tfsdk:"eval_template_id"`
	Enabled               types.Bool    `tfsdk:"enabled"`
	TargetType            types.String  `tfsdk:"target_type"`
	SamplingRate          types.Float64 `tfsdk:"sampling_rate"`
	MaxInvocationsPerHour types.Int64   `tfsdk:"max_invocations_per_hour"`
	Filters               types.String  `tfsdk:"filters"`
	VariableMapping       types.String  `tfsdk:"variable_mapping"`
	LLMConnectionID       types.String  `tfsdk:"llm_connection_id"`
	ModelParams           types.String  `tfsdk:"model_params"`
	DependsOnRuleIDs      types.List    `tfsdk:"depends_on_rule_ids"`
	DatasetID             types.String  `tfsdk:"dataset_id"`
}

func (m *genaiEvaluatorResourceModel) GetID() types.String {
	return m.ID
}

func (m *genaiEvaluatorResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *genaiEvaluatorResourceModel) FromClientModel(
	ctx context.Context,
	model *clientmodels.GenAIEvaluationRule,
	diagnosticsOut *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.EvalTemplateID = types.StringValue(model.EvaluatorID)
	m.TargetType = types.StringValue(model.TargetType)

	if model.Enabled != nil {
		m.Enabled = types.BoolValue(*model.Enabled)
	} else {
		m.Enabled = types.BoolValue(false)
	}

	if model.SamplingRate != nil {
		m.SamplingRate = types.Float64Value(*model.SamplingRate)
	} else {
		m.SamplingRate = types.Float64Value(0)
	}

	if model.MaxInvocationsPerHour != nil {
		m.MaxInvocationsPerHour = types.Int64Value(
			*model.MaxInvocationsPerHour,
		)
	} else {
		m.MaxInvocationsPerHour = types.Int64Value(0)
	}

	m.LLMConnectionID = optionalString(
		derefString(model.LLMConnectionID), m.LLMConnectionID,
	)
	m.DatasetID = optionalString(model.DatasetID, m.DatasetID)

	var dependsOn []string
	if model.DependsOnRuleIDs != nil {
		dependsOn = *model.DependsOnRuleIDs
	}
	m.DependsOnRuleIDs = resourceutils.SliceToStringList(
		ctx, dependsOn, m.DependsOnRuleIDs, diagnosticsOut,
	)

	m.Filters = resourceutils.RawToJSONString(model.Filters, m.Filters)
	m.VariableMapping = resourceutils.RawToJSONString(
		model.VariableMapping, m.VariableMapping,
	)
	m.ModelParams = resourceutils.RawToJSONString(
		model.ModelParams, m.ModelParams,
	)
}

func (m *genaiEvaluatorResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.GenAIEvaluationRule,
) error {
	model.ID = m.ID.ValueString()
	model.Name = m.Name.ValueString()
	model.EvaluatorID = m.EvalTemplateID.ValueString()
	model.TargetType = m.TargetType.ValueString()
	model.DatasetID = m.DatasetID.ValueString()

	// The update endpoint merges rather than replaces, so every
	// attribute Terraform owns is sent on each apply. Without this a
	// rule could never be disabled or have its sampling rate lowered
	// back to zero.
	enabled := m.Enabled.ValueBool()
	model.Enabled = &enabled

	samplingRate := m.SamplingRate.ValueFloat64()
	model.SamplingRate = &samplingRate

	maxInvocations := m.MaxInvocationsPerHour.ValueInt64()
	model.MaxInvocationsPerHour = &maxInvocations

	llmConnectionID := m.LLMConnectionID.ValueString()
	model.LLMConnectionID = &llmConnectionID

	dependsOn, err := resourceutils.StringListToSlice(
		ctx, m.DependsOnRuleIDs,
	)
	if err != nil {
		return err
	}
	if dependsOn == nil {
		dependsOn = []string{}
	}
	model.DependsOnRuleIDs = &dependsOn

	filters, err := resourceutils.JSONStringToRaw(m.Filters, "filters")
	if err != nil {
		return err
	}
	model.Filters = filters

	variableMapping, err := resourceutils.JSONStringToRaw(
		m.VariableMapping, "variable_mapping",
	)
	if err != nil {
		return err
	}
	model.VariableMapping = variableMapping

	modelParams, err := resourceutils.JSONStringToRaw(
		m.ModelParams, "model_params",
	)
	if err != nil {
		return err
	}
	model.ModelParams = modelParams

	return nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

// optionalString keeps an unset optional attribute null rather than
// turning it into "" when the API omits the field.
func optionalString(value string, prior types.String) types.String {
	if value == "" {
		if prior.IsNull() || prior.IsUnknown() {
			return types.StringNull()
		}

		return types.StringNull()
	}

	return types.StringValue(value)
}
