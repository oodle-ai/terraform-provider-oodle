package syntheticmonitor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/resourceutils"
	"terraform-provider-oodle/internal/validatorutils"
)

type syntheticMonitorResourceModel struct {
	ID         types.String                 `tfsdk:"id"`
	Name       types.String                 `tfsdk:"name"`
	Enabled    types.Bool                   `tfsdk:"enabled"`
	RuleType   types.String                 `tfsdk:"rule_type"`
	RuleConfig *ruleConfigModel             `tfsdk:"rule_config"`
	Interval   validatorutils.DurationValue `tfsdk:"interval"`
	Timeout    validatorutils.DurationValue `tfsdk:"timeout"`
}

type ruleConfigModel struct {
	HTTP      *httpConfigModel      `tfsdk:"http"`
	Multistep *multistepConfigModel `tfsdk:"multistep"`
}

// httpConfigModel is shared by the single-step "http" rule config and by each
// step's request in a multi-step monitor.
type httpConfigModel struct {
	URL                 types.String      `tfsdk:"url"`
	Method              types.String      `tfsdk:"method"`
	Headers             map[string]string `tfsdk:"headers"`
	Body                types.String      `tfsdk:"body"`
	ExpectedStatusCodes []types.String    `tfsdk:"expected_status_codes"`
	ExcludedStatusCodes []types.String    `tfsdk:"excluded_status_codes"`
	ExpectedBody        types.String      `tfsdk:"expected_body"`
	MaxResponseTimeMs   types.Int64       `tfsdk:"max_response_time_ms"`
	ExpectedHeaders     map[string]string `tfsdk:"expected_headers"`
	FollowRedirects     types.Bool        `tfsdk:"follow_redirects"`
	InsecureSkipVerify  types.Bool        `tfsdk:"insecure_skip_verify"`
	BasicAuth           *basicAuthModel   `tfsdk:"basic_auth"`
	BearerToken         types.String      `tfsdk:"bearer_token"`
}

type basicAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type multistepConfigModel struct {
	Steps []stepModel `tfsdk:"steps"`
}

type stepModel struct {
	Name              types.String       `tfsdk:"name"`
	Request           *httpConfigModel   `tfsdk:"request"`
	Extract           []extractRuleModel `tfsdk:"extract"`
	ContinueOnFailure types.Bool         `tfsdk:"continue_on_failure"`
	ExitOnSuccess     types.Bool         `tfsdk:"exit_on_success"`
}

type extractRuleModel struct {
	Name   types.String `tfsdk:"name"`
	Source types.String `tfsdk:"source"`
	Parser types.String `tfsdk:"parser"`
	Query  types.String `tfsdk:"query"`
	Secret types.Bool   `tfsdk:"secret"`
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
	m.Interval = validatorutils.NewDurationValue(model.Interval)
	m.Timeout = validatorutils.NewDurationValue(model.Timeout)

	m.RuleConfig = &ruleConfigModel{}
	if model.RuleConfig.HTTP != nil {
		m.RuleConfig.HTTP = httpConfigFromClientModel(model.RuleConfig.HTTP)
	}
	if model.RuleConfig.Multistep != nil {
		steps := make([]stepModel, len(model.RuleConfig.Multistep.Steps))
		for i, step := range model.RuleConfig.Multistep.Steps {
			s := stepModel{
				Name:              types.StringValue(step.Name),
				Request:           httpConfigFromClientModel(&step.Request),
				ContinueOnFailure: types.BoolValue(step.ContinueOnFailure),
				ExitOnSuccess:     types.BoolValue(step.ExitOnSuccess),
			}
			if len(step.Extract) > 0 {
				s.Extract = make([]extractRuleModel, len(step.Extract))
				for j, ex := range step.Extract {
					s.Extract[j] = extractRuleModel{
						Name:   types.StringValue(ex.Name),
						Source: types.StringValue(ex.Source),
						Parser: types.StringValue(ex.Parser),
						Query:  types.StringValue(ex.Query),
						Secret: types.BoolValue(ex.Secret),
					}
				}
			}
			steps[i] = s
		}
		m.RuleConfig.Multistep = &multistepConfigModel{Steps: steps}
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

	if m.RuleConfig == nil {
		return nil
	}

	if m.RuleConfig.HTTP != nil {
		model.RuleConfig.HTTP = httpConfigToClientModel(m.RuleConfig.HTTP)
	}
	if m.RuleConfig.Multistep != nil {
		steps := make([]clientmodels.SyntheticMonitorStep, len(m.RuleConfig.Multistep.Steps))
		for i, step := range m.RuleConfig.Multistep.Steps {
			s := clientmodels.SyntheticMonitorStep{
				Name:              step.Name.ValueString(),
				ContinueOnFailure: step.ContinueOnFailure.ValueBool(),
				ExitOnSuccess:     step.ExitOnSuccess.ValueBool(),
			}
			if step.Request != nil {
				s.Request = *httpConfigToClientModel(step.Request)
			}
			if len(step.Extract) > 0 {
				s.Extract = make([]clientmodels.SyntheticMonitorExtractRule, len(step.Extract))
				for j, ex := range step.Extract {
					s.Extract[j] = clientmodels.SyntheticMonitorExtractRule{
						Name:   ex.Name.ValueString(),
						Source: ex.Source.ValueString(),
						Parser: ex.Parser.ValueString(),
						Query:  ex.Query.ValueString(),
						Secret: ex.Secret.ValueBool(),
					}
				}
			}
			steps[i] = s
		}
		model.RuleConfig.Multistep = &clientmodels.SyntheticMonitorMultistepConfig{Steps: steps}
	}

	return nil
}

// httpConfigFromClientModel converts a client HTTP config into the TF model,
// leaving optional/empty fields null so they round-trip cleanly.
func httpConfigFromClientModel(c *clientmodels.SyntheticMonitorHTTPConfig) *httpConfigModel {
	cfg := &httpConfigModel{
		URL:                types.StringValue(c.URL),
		Method:             types.StringValue(c.Method),
		FollowRedirects:    types.BoolValue(c.FollowRedirects),
		InsecureSkipVerify: types.BoolValue(c.InsecureSkipVerify),
	}

	if c.Body != "" {
		cfg.Body = types.StringValue(c.Body)
	}
	if len(c.Headers) > 0 {
		cfg.Headers = c.Headers
	}
	if len(c.ExpectedStatusCodes) > 0 {
		cfg.ExpectedStatusCodes = stringsToTFList(c.ExpectedStatusCodes)
	}
	if len(c.ExcludedStatusCodes) > 0 {
		cfg.ExcludedStatusCodes = stringsToTFList(c.ExcludedStatusCodes)
	}
	if c.ExpectedBody != "" {
		cfg.ExpectedBody = types.StringValue(c.ExpectedBody)
	}
	if c.MaxResponseTimeMs != 0 {
		cfg.MaxResponseTimeMs = types.Int64Value(c.MaxResponseTimeMs)
	}
	if len(c.ExpectedHeaders) > 0 {
		cfg.ExpectedHeaders = c.ExpectedHeaders
	}
	if c.BasicAuth != nil {
		cfg.BasicAuth = &basicAuthModel{
			Username: types.StringValue(c.BasicAuth.Username),
			Password: types.StringValue(c.BasicAuth.Password),
		}
	}
	if c.BearerToken != "" {
		cfg.BearerToken = types.StringValue(c.BearerToken)
	}

	return cfg
}

// httpConfigToClientModel converts a TF HTTP config model into the client model.
func httpConfigToClientModel(m *httpConfigModel) *clientmodels.SyntheticMonitorHTTPConfig {
	cfg := &clientmodels.SyntheticMonitorHTTPConfig{
		URL:                m.URL.ValueString(),
		Method:             m.Method.ValueString(),
		FollowRedirects:    m.FollowRedirects.ValueBool(),
		InsecureSkipVerify: m.InsecureSkipVerify.ValueBool(),
	}

	if !m.Body.IsNull() && !m.Body.IsUnknown() {
		cfg.Body = m.Body.ValueString()
	}
	if len(m.Headers) > 0 {
		cfg.Headers = m.Headers
	}
	if len(m.ExpectedStatusCodes) > 0 {
		cfg.ExpectedStatusCodes = tfListToStrings(m.ExpectedStatusCodes)
	}
	if len(m.ExcludedStatusCodes) > 0 {
		cfg.ExcludedStatusCodes = tfListToStrings(m.ExcludedStatusCodes)
	}
	if !m.ExpectedBody.IsNull() && !m.ExpectedBody.IsUnknown() {
		cfg.ExpectedBody = m.ExpectedBody.ValueString()
	}
	if !m.MaxResponseTimeMs.IsNull() && !m.MaxResponseTimeMs.IsUnknown() {
		cfg.MaxResponseTimeMs = m.MaxResponseTimeMs.ValueInt64()
	}
	if len(m.ExpectedHeaders) > 0 {
		cfg.ExpectedHeaders = m.ExpectedHeaders
	}
	if m.BasicAuth != nil {
		cfg.BasicAuth = &clientmodels.SyntheticMonitorBasicAuth{
			Username: m.BasicAuth.Username.ValueString(),
			Password: m.BasicAuth.Password.ValueString(),
		}
	}
	if !m.BearerToken.IsNull() && !m.BearerToken.IsUnknown() {
		cfg.BearerToken = m.BearerToken.ValueString()
	}

	return cfg
}

func stringsToTFList(in []string) []types.String {
	out := make([]types.String, len(in))
	for i, s := range in {
		out[i] = types.StringValue(s)
	}
	return out
}

func tfListToStrings(in []types.String) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = s.ValueString()
	}
	return out
}
