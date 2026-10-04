package genaievaluator

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

func newEvaluatorModel() *genaiEvaluatorResourceModel {
	return &genaiEvaluatorResourceModel{
		Name:             types.StringValue("tone"),
		EvalTemplateID:   types.StringValue("template-1"),
		DependsOnRuleIDs: types.ListNull(types.StringType),
		Filters:          resourceutils.NewJSONNull(),
		VariableMapping:  resourceutils.NewJSONNull(),
		ModelParams:      resourceutils.NewJSONNull(),
		Params:           resourceutils.NewJSONNull(),
	}
}

// Unset params are sent as an empty object, because the update
// endpoint keeps the stored values when the request has none.
func TestEvaluatorUnsetParamsSendEmptyObject(t *testing.T) {
	written := &clientmodels.GenAIEvaluationRule{}
	assert.Nil(t, newEvaluatorModel().ToClientModel(context.Background(), written))
	assert.Equal(t, "{}", string(written.Params))
}

func TestEvaluatorParamsRoundTrip(t *testing.T) {
	ctx := context.Background()
	m := newEvaluatorModel()
	m.Params = resourceutils.NewJSONValue(`{"threshold": 0.8, "words": ["a"]}`)

	written := &clientmodels.GenAIEvaluationRule{}
	assert.Nil(t, m.ToClientModel(ctx, written))
	assert.Equal(t, `{"threshold": 0.8, "words": ["a"]}`, string(written.Params))

	out := &diag.Diagnostics{}
	m.FromClientModel(ctx, &clientmodels.GenAIEvaluationRule{
		ID:                "rule-1",
		Name:              "tone",
		EvaluatorID:       "template-1",
		Params:            json.RawMessage(`{"words":["a"],"threshold":0.8}`),
		ScoreInputRuleIDs: []string{"rule-0"},
	}, out)
	assert.False(t, out.HasError())
	// The configured text is kept, not the API's spelling of it.
	assert.Equal(
		t, `{"threshold": 0.8, "words": ["a"]}`, m.Params.ValueString(),
	)
	assert.Equal(t, 1, len(m.ScoreInputRuleIDs.Elements()))
}

// An empty object, or settings set to null, are stored as no params.
// The configuration is kept, so the apply does not report an
// inconsistent result.
func TestEvaluatorEmptyParamsStayAsConfigured(t *testing.T) {
	for _, configured := range []string{`{}`, `{"threshold": null}`} {
		m := newEvaluatorModel()
		m.Params = resourceutils.NewJSONValue(configured)
		out := &diag.Diagnostics{}
		m.FromClientModel(context.Background(), &clientmodels.GenAIEvaluationRule{
			ID: "rule-1",
		}, out)
		assert.False(t, out.HasError())
		assert.Equal(t, configured, m.Params.ValueString())
		// No score inputs is an empty list, not unknown.
		assert.False(t, m.ScoreInputRuleIDs.IsUnknown())
		assert.Equal(t, 0, len(m.ScoreInputRuleIDs.Elements()))
	}
}

// A value changed outside Terraform shows up.
func TestEvaluatorChangedParamsAreRead(t *testing.T) {
	m := newEvaluatorModel()
	m.Params = resourceutils.NewJSONValue(`{"threshold": 0.8}`)
	m.FromClientModel(context.Background(), &clientmodels.GenAIEvaluationRule{
		ID:     "rule-1",
		Params: json.RawMessage(`{"threshold":0.5}`),
	}, &diag.Diagnostics{})
	assert.Equal(t, `{"threshold":0.5}`, m.Params.ValueString())
}

func TestEvaluatorNoParamsStaysNull(t *testing.T) {
	m := newEvaluatorModel()
	m.FromClientModel(context.Background(), &clientmodels.GenAIEvaluationRule{
		ID: "rule-1",
	}, &diag.Diagnostics{})
	assert.True(t, m.Params.IsNull())
}
