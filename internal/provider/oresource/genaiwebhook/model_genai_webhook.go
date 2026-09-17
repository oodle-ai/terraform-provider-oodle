package genaiwebhook

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

type genaiWebhookResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	URL             types.String `tfsdk:"url"`
	Headers         types.Map    `tfsdk:"headers"`
	TimeoutSeconds  types.Int64  `tfsdk:"timeout_seconds"`
	RequestTemplate types.String `tfsdk:"request_template"`
	OutputPath      types.String `tfsdk:"output_path"`
}

func (m *genaiWebhookResourceModel) GetID() types.String {
	return m.ID
}

func (m *genaiWebhookResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *genaiWebhookResourceModel) FromClientModel(
	_ context.Context,
	model *clientmodels.GenAIWebhook,
	_ *diag.Diagnostics,
) {
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.URL = types.StringValue(model.URL)
	m.Description = optionalString(model.Description)

	// headers are write-only: the API stores them encrypted and
	// returns only their names, so the values already held are kept
	// rather than being cleared.

	m.TimeoutSeconds = types.Int64Value(int64(model.TimeoutSeconds))
	// Both are computed: the server stores "" for the default and
	// the form reads "" as the default too, so "" is a real value
	// here rather than an unset one.
	m.RequestTemplate = types.StringValue(model.RequestTemplate)
	m.OutputPath = types.StringValue(model.OutputPath)
}

func (m *genaiWebhookResourceModel) ToClientModel(
	ctx context.Context,
	model *clientmodels.GenAIWebhook,
) error {
	model.ID = m.ID.ValueString()
	model.Name = m.Name.ValueString()
	model.Description = m.Description.ValueString()
	model.URL = m.URL.ValueString()
	model.TimeoutSeconds = int(m.TimeoutSeconds.ValueInt64())
	model.RequestTemplate = m.RequestTemplate.ValueString()
	model.OutputPath = m.OutputPath.ValueString()

	// Always sent, empty included: the update endpoint keeps the
	// stored headers when the request carries none, and a
	// configuration that removed them means "none".
	headers := map[string]string{}
	if !m.Headers.IsNull() && !m.Headers.IsUnknown() {
		diags := m.Headers.ElementsAs(ctx, &headers, false)
		if diags.HasError() {
			return fmt.Errorf("failed to read headers: %v", diags.Errors())
		}
	}
	model.Headers = headers

	return nil
}

// optionalString keeps an unset optional attribute null instead of
// turning it into "" when the API omits the field.
func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}
