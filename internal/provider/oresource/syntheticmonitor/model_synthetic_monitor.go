package syntheticmonitor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
)

type syntheticMonitorResourceModel struct {
	ID         types.String     `tfsdk:"id"`
	Name       types.String     `tfsdk:"name"`
	Enabled    types.Bool       `tfsdk:"enabled"`
	RuleType   types.String     `tfsdk:"rule_type"`
	RuleConfig *ruleConfigModel `tfsdk:"rule_config"`
	Interval   types.String     `tfsdk:"interval"`
	Timeout    types.String     `tfsdk:"timeout"`
}

type ruleConfigModel struct {
	HTTP *httpConfigModel `tfsdk:"http"`
}

type httpConfigModel struct {
	URL                 types.String      `tfsdk:"url"`
	Method              types.String      `tfsdk:"method"`
	Headers             map[string]string `tfsdk:"headers"`
	Body                types.String      `tfsdk:"body"`
	ExpectedStatusCodes []types.String    `tfsdk:"expected_status_codes"`
	FollowRedirects     types.Bool        `tfsdk:"follow_redirects"`
	InsecureSkipVerify  types.Bool        `tfsdk:"insecure_skip_verify"`
}

var _ resourceutils.ResourceModel[*clientmodels.SyntheticMonitor] = (*syntheticMonitorResourceModel)(nil)

func (m *syntheticMonitorResourceModel) GetID() types.String {
	return m.ID
}

func (m *syntheticMonitorResourceModel) SetID(id types.String) {
	m.ID = id
}

func (m *syntheticMonitorResourceModel) FromClientModel(
	_ context.Context,
	model *clientmodels.SyntheticMonitor,
	_ *diag.Diagnostics,
) {
	// Reset the model to clear any existing data.
	*m = syntheticMonitorResourceModel{}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.Enabled = types.BoolValue(model.Enabled)
	m.RuleType = types.StringValue(model.RuleType)
	m.Interval = types.StringValue(model.Interval)
	m.Timeout = types.StringValue(model.Timeout)

	m.RuleConfig = &ruleConfigModel{}
	if model.RuleConfig.HTTP != nil {
		httpCfg := &httpConfigModel{
			URL:                types.StringValue(model.RuleConfig.HTTP.URL),
			Method:             types.StringValue(model.RuleConfig.HTTP.Method),
			FollowRedirects:    types.BoolValue(model.RuleConfig.HTTP.FollowRedirects),
			InsecureSkipVerify: types.BoolValue(model.RuleConfig.HTTP.InsecureSkipVerify),
		}

		if model.RuleConfig.HTTP.Body != "" {
			httpCfg.Body = types.StringValue(model.RuleConfig.HTTP.Body)
		}

		if len(model.RuleConfig.HTTP.Headers) > 0 {
			httpCfg.Headers = model.RuleConfig.HTTP.Headers
		}

		if len(model.RuleConfig.HTTP.ExpectedStatusCodes) > 0 {
			httpCfg.ExpectedStatusCodes = make([]types.String, len(model.RuleConfig.HTTP.ExpectedStatusCodes))
			for i, code := range model.RuleConfig.HTTP.ExpectedStatusCodes {
				httpCfg.ExpectedStatusCodes[i] = types.StringValue(code)
			}
		}

		m.RuleConfig.HTTP = httpCfg
	}
}

func (m *syntheticMonitorResourceModel) ToClientModel(
	_ context.Context,
	model *clientmodels.SyntheticMonitor,
) error {
	if !m.ID.IsNull() && !m.ID.IsUnknown() {
		model.ID = m.ID.ValueString()
	}

	model.Name = m.Name.ValueString()
	model.Enabled = m.Enabled.ValueBool()
	model.RuleType = m.RuleType.ValueString()
	model.Interval = m.Interval.ValueString()
	model.Timeout = m.Timeout.ValueString()

	if m.RuleConfig != nil && m.RuleConfig.HTTP != nil {
		httpCfg := &clientmodels.SyntheticMonitorHTTPConfig{
			URL:                m.RuleConfig.HTTP.URL.ValueString(),
			Method:             m.RuleConfig.HTTP.Method.ValueString(),
			FollowRedirects:    m.RuleConfig.HTTP.FollowRedirects.ValueBool(),
			InsecureSkipVerify: m.RuleConfig.HTTP.InsecureSkipVerify.ValueBool(),
		}

		if !m.RuleConfig.HTTP.Body.IsNull() && !m.RuleConfig.HTTP.Body.IsUnknown() {
			httpCfg.Body = m.RuleConfig.HTTP.Body.ValueString()
		}

		if len(m.RuleConfig.HTTP.Headers) > 0 {
			httpCfg.Headers = m.RuleConfig.HTTP.Headers
		}

		if len(m.RuleConfig.HTTP.ExpectedStatusCodes) > 0 {
			httpCfg.ExpectedStatusCodes = make([]string, len(m.RuleConfig.HTTP.ExpectedStatusCodes))
			for i, code := range m.RuleConfig.HTTP.ExpectedStatusCodes {
				httpCfg.ExpectedStatusCodes[i] = code.ValueString()
			}
		}

		model.RuleConfig = clientmodels.SyntheticMonitorRuleConfig{
			HTTP: httpCfg,
		}
	}

	return nil
}
