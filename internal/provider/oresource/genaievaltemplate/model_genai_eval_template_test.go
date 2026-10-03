package genaievaltemplate

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

func newTemplateModel(templateType string) *genaiEvalTemplateResourceModel {
	return &genaiEvalTemplateResourceModel{
		Name:         types.StringValue("tone"),
		Type:         types.StringValue(templateType),
		Vars:         types.ListNull(types.StringType),
		OutputSchema: resourceutils.NewJSONNull(),
		ModelParams:  resourceutils.NewJSONNull(),
		Params:       resourceutils.NewJSONNull(),
		LibraryPins:  types.MapNull(types.Int64Type),
	}
}

// A code template always sends params and pins, so that removing
// them from the configuration clears them.
func TestCodeTemplateSendsEmptyParamsAndPins(t *testing.T) {
	written := &clientmodels.GenAIEvalTemplate{}
	assert.Nil(t, newTemplateModel("code").ToClientModel(context.Background(), written))
	assert.Equal(t, "[]", string(written.Params))
	assert.NotNil(t, written.LibraryPins)
	assert.Equal(t, 0, len(*written.LibraryPins))

	encoded, err := json.Marshal(written)
	assert.Nil(t, err)
	assert.Contains(t, `"libraryPins":{}`, string(encoded))
}

// Any other type sends neither: the API refuses them.
func TestLLMTemplateSendsNoParamsOrPins(t *testing.T) {
	written := &clientmodels.GenAIEvalTemplate{}
	assert.Nil(t, newTemplateModel("llm").ToClientModel(context.Background(), written))
	assert.Nil(t, written.Params)
	assert.Nil(t, written.LibraryPins)

	m := newTemplateModel("llm")
	m.Params = resourceutils.NewJSONValue(`[]`)
	assert.NotNil(t, m.ToClientModel(context.Background(), written))
}

func TestCodeTemplateRoundTrip(t *testing.T) {
	ctx := context.Background()
	configured := `[{"name": "threshold", "type": "number", ` +
		`"default": 0.5, "required": false, "label": ""}]`

	m := newTemplateModel("code")
	m.Params = resourceutils.NewJSONValue(configured)
	pins, diags := types.MapValueFrom(
		ctx, types.Int64Type, map[string]int64{"helpers": 2},
	)
	assert.False(t, diags.HasError())
	m.LibraryPins = pins

	written := &clientmodels.GenAIEvalTemplate{}
	assert.Nil(t, m.ToClientModel(ctx, written))
	assert.Equal(t, int64(2), (*written.LibraryPins)["helpers"])

	// The API drops fields that hold their zero value.
	answered := map[string]int64{"helpers": 2}
	out := &diag.Diagnostics{}
	m.FromClientModel(ctx, &clientmodels.GenAIEvalTemplate{
		ID:          "template-1",
		Name:        "tone",
		Type:        "code",
		Params:      json.RawMessage(`[{"name":"threshold","type":"number","default":0.5}]`),
		LibraryPins: &answered,
		Version:     1,
	}, out)
	assert.False(t, out.HasError())
	assert.Equal(t, configured, m.Params.ValueString())
	assert.Equal(t, 1, len(m.LibraryPins.Elements()))
}

// An empty array configured stays as configured although the API
// gives back nothing; unset stays null.
func TestCodeTemplateEmptyParams(t *testing.T) {
	m := newTemplateModel("code")
	m.Params = resourceutils.NewJSONValue(`[]`)
	m.FromClientModel(context.Background(), &clientmodels.GenAIEvalTemplate{
		ID: "template-1", Type: "code",
	}, &diag.Diagnostics{})
	assert.Equal(t, `[]`, m.Params.ValueString())
	assert.True(t, m.LibraryPins.IsNull())

	m = newTemplateModel("code")
	m.FromClientModel(context.Background(), &clientmodels.GenAIEvalTemplate{
		ID: "template-1", Type: "code",
	}, &diag.Diagnostics{})
	assert.True(t, m.Params.IsNull())
}

func TestSameParamSpecsSeesRealChanges(t *testing.T) {
	assert.False(t, sameParamSpecs(
		[]byte(`[{"name":"a","type":"number","default":1}]`),
		[]byte(`[{"name":"a","type":"number","default":2}]`),
	))
	assert.True(t, sameParamSpecs(
		[]byte(`[{"name":"a","type":"number","default":null}]`),
		[]byte(`[{"name":"a","type":"number"}]`),
	))
}
