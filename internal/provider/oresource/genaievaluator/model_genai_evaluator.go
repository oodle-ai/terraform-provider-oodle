package genaievaluator

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type genaiEvaluatorResourceModel struct {
	ID                    types.String       `tfsdk:"id"`
	Name                  types.String       `tfsdk:"name"`
	EvalTemplateID        types.String       `tfsdk:"eval_template_id"`
	Enabled               types.Bool         `tfsdk:"enabled"`
	TargetType            types.String       `tfsdk:"target_type"`
	SamplingRate          types.Float64      `tfsdk:"sampling_rate"`
	MaxInvocationsPerHour types.Int64        `tfsdk:"max_invocations_per_hour"`
	Filters               resourceutils.JSON `tfsdk:"filters"`
	VariableMapping       resourceutils.JSON `tfsdk:"variable_mapping"`
	LLMConnectionID       types.String       `tfsdk:"llm_connection_id"`
	ModelParams           resourceutils.JSON `tfsdk:"model_params"`
	Params                resourceutils.JSON `tfsdk:"params"`
	ScoreInputRuleIDs     types.List         `tfsdk:"score_input_rule_ids"`
	DependsOnRuleIDs      types.List         `tfsdk:"depends_on_rule_ids"`
	DatasetID             types.String       `tfsdk:"dataset_id"`
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

	m.Filters = resourceutils.RawToJSON(model.Filters)
	m.VariableMapping = resourceutils.RawToJSON(model.VariableMapping)
	m.ModelParams = resourceutils.RawToJSON(model.ModelParams)
	m.Params = resourceutils.RawToJSONKeepingPrior(
		model.Params, m.Params, sameParamValues,
	)

	scoreInputs, diags := types.ListValueFrom(
		ctx, types.StringType, nonNilStrings(model.ScoreInputRuleIDs),
	)
	diagnosticsOut.Append(diags...)
	m.ScoreInputRuleIDs = scoreInputs
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

	filters, err := resourceutils.JSONToRaw(m.Filters, "filters")
	if err != nil {
		return err
	}
	model.Filters = filters

	variableMapping, err := resourceutils.JSONToRaw(
		m.VariableMapping, "variable_mapping",
	)
	if err != nil {
		return err
	}
	model.VariableMapping = variableMapping

	modelParams, err := resourceutils.JSONToRaw(
		m.ModelParams, "model_params",
	)
	if err != nil {
		return err
	}
	model.ModelParams = modelParams

	params, err := resourceutils.JSONToRaw(m.Params, "params")
	if err != nil {
		return err
	}
	// The update endpoint keeps the stored values when the request
	// has none, so an empty object is sent to clear them. The API
	// accepts an empty object on every template type.
	if params == nil {
		params = json.RawMessage("{}")
	}
	model.Params = params

	return nil
}

// sameParamValues reports whether two params objects hold the same
// values. The API stores no params for an empty object, and a setting
// set to null means "not set", so both read as absent.
func sameParamValues(a, b []byte) bool {
	canonical := func(raw []byte) (map[string]any, bool) {
		out := map[string]any{}
		if len(raw) == 0 || string(raw) == "null" {
			return out, true
		}

		var values map[string]any
		if json.Unmarshal(raw, &values) != nil {
			return nil, false
		}
		for name, value := range values {
			if value != nil {
				out[name] = value
			}
		}

		return out, true
	}

	left, ok := canonical(a)
	if !ok {
		return false
	}
	right, ok := canonical(b)
	if !ok {
		return false
	}

	return reflect.DeepEqual(left, right)
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}

	return values
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
