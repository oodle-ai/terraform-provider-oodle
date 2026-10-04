package genaievaltemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	Params             resourceutils.JSON `tfsdk:"params"`
	LibraryPins        types.Map          `tfsdk:"library_pins"`
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
	m.Params = resourceutils.RawToJSONKeepingPrior(
		model.Params, m.Params, sameParamSpecs,
	)

	var pins map[string]int64
	if model.LibraryPins != nil {
		pins = *model.LibraryPins
	}
	m.LibraryPins = pinsToMap(ctx, pins, m.LibraryPins, diagnosticsOut)
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

	params, err := resourceutils.JSONToRaw(m.Params, "params")
	if err != nil {
		return err
	}

	pins := map[string]int64{}
	if !m.LibraryPins.IsNull() && !m.LibraryPins.IsUnknown() {
		if diags := m.LibraryPins.ElementsAs(ctx, &pins, false); diags.HasError() {
			return fmt.Errorf(
				"failed to read library_pins: %v", diags.Errors(),
			)
		}
	}

	// The API refuses params and library pins on any template type but
	// 'code'. On a code template both are always sent, because the
	// update endpoint keeps the stored value when the request has none:
	// removing them from the configuration must clear them.
	if model.Type == codeTemplateType {
		if params == nil {
			params = json.RawMessage("[]")
		}
		model.Params = params
		model.LibraryPins = &pins
	} else {
		if params != nil || len(pins) > 0 {
			return fmt.Errorf(
				"params and library_pins are only for 'code' templates",
			)
		}
		model.Params = nil
		model.LibraryPins = nil
	}

	return nil
}

const codeTemplateType = "code"

// paramSpec is one setting of a code template, in the shape the API
// stores it. The API drops a field that holds its zero value, such as
// "required": false, and any field it does not know.
type paramSpec struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Label       string          `json:"label,omitempty"`
	Description string          `json:"description,omitempty"`
	Default     json.RawMessage `json:"default,omitempty"`
	Required    bool            `json:"required,omitempty"`
	Options     []string        `json:"options,omitempty"`
}

// sameParamSpecs reports whether two params arrays declare the same
// settings once both are in the form the API stores. Without this, a
// configuration that writes "required": false would differ from what
// the API gives back on every plan.
func sameParamSpecs(a, b []byte) bool {
	canonical := func(raw []byte) (any, bool) {
		var specs []paramSpec
		if len(raw) > 0 && string(raw) != "null" {
			if json.Unmarshal(raw, &specs) != nil {
				return nil, false
			}
		}
		for i := range specs {
			if string(specs[i].Default) == "null" {
				specs[i].Default = nil
			}
		}
		if specs == nil {
			specs = []paramSpec{}
		}

		encoded, err := json.Marshal(specs)
		if err != nil {
			return nil, false
		}

		var out any
		if json.Unmarshal(encoded, &out) != nil {
			return nil, false
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

// pinsToMap converts library pins from the API into a Terraform map. No
// pins become a null map when the prior value was null, so that an unset
// attribute does not start to show a diff against an empty map.
func pinsToMap(
	ctx context.Context,
	pins map[string]int64,
	prior types.Map,
	diagnosticsOut *diag.Diagnostics,
) types.Map {
	if len(pins) == 0 {
		if prior.IsNull() || prior.IsUnknown() {
			return types.MapNull(types.Int64Type)
		}

		return types.MapValueMust(types.Int64Type, map[string]attr.Value{})
	}

	result, diags := types.MapValueFrom(ctx, types.Int64Type, pins)
	if diags.HasError() {
		diagnosticsOut.Append(diags...)
		return prior
	}

	return result
}

// optionalString keeps an unset optional attribute null rather than
// turning it into "" when the API omits the field.
func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}
