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
	HTTP       *httpConfigModel       `tfsdk:"http"`
	Ping       *pingConfigModel       `tfsdk:"ping"`
	DNS        *dnsConfigModel        `tfsdk:"dns"`
	TCP        *tcpConfigModel        `tfsdk:"tcp"`
	Traceroute *tracerouteConfigModel `tfsdk:"traceroute"`
	SSL        *sslConfigModel        `tfsdk:"ssl"`
	Multistep  *multistepConfigModel  `tfsdk:"multistep"`
}

type pingConfigModel struct {
	Host       types.String `tfsdk:"host"`
	Count      types.Int64  `tfsdk:"count"`
	IntervalMs types.Int64  `tfsdk:"interval_ms"`
}

type dnsConfigModel struct {
	Domain           types.String   `tfsdk:"domain"`
	RecordType       types.String   `tfsdk:"record_type"`
	ExpectedValues   []types.String `tfsdk:"expected_values"`
	Nameserver       types.String   `tfsdk:"nameserver"`
	ExpectResolution types.Bool     `tfsdk:"expect_resolution"`
}

type tcpConfigModel struct {
	Host types.String `tfsdk:"host"`
	Port types.Int64  `tfsdk:"port"`
}

type tracerouteConfigModel struct {
	Host            types.String `tfsdk:"host"`
	MaxHops         types.Int64  `tfsdk:"max_hops"`
	TimeoutPerHopMs types.Int64  `tfsdk:"timeout_per_hop_ms"`
}

type sslConfigModel struct {
	Host                      types.String `tfsdk:"host"`
	Port                      types.Int64  `tfsdk:"port"`
	WarnDaysBeforeExpiry      types.Int64  `tfsdk:"warn_days_before_expiry"`
	CriticalDaysBeforeExpiry  types.Int64  `tfsdk:"critical_days_before_expiry"`
	InsecureSkipVerify        types.Bool   `tfsdk:"insecure_skip_verify"`
	CheckCertificateAuthority types.Bool   `tfsdk:"check_certificate_authority"`
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
	if c := model.RuleConfig.Ping; c != nil {
		m.RuleConfig.Ping = &pingConfigModel{
			Host:       types.StringValue(c.Host),
			Count:      optionalInt64(c.Count),
			IntervalMs: optionalInt64(c.IntervalMs),
		}
	}
	if c := model.RuleConfig.DNS; c != nil {
		cfg := &dnsConfigModel{
			Domain:           types.StringValue(c.Domain),
			RecordType:       optionalString(c.RecordType),
			Nameserver:       optionalString(c.Nameserver),
			ExpectResolution: types.BoolValue(c.ExpectResolution),
		}
		if len(c.ExpectedValues) > 0 {
			cfg.ExpectedValues = stringsToTFList(c.ExpectedValues)
		}
		m.RuleConfig.DNS = cfg
	}
	if c := model.RuleConfig.TCP; c != nil {
		m.RuleConfig.TCP = &tcpConfigModel{
			Host: types.StringValue(c.Host),
			Port: types.Int64Value(c.Port),
		}
	}
	if c := model.RuleConfig.Traceroute; c != nil {
		m.RuleConfig.Traceroute = &tracerouteConfigModel{
			Host:            types.StringValue(c.Host),
			MaxHops:         optionalInt64(c.MaxHops),
			TimeoutPerHopMs: optionalInt64(c.TimeoutPerHopMs),
		}
	}
	if c := model.RuleConfig.SSL; c != nil {
		m.RuleConfig.SSL = &sslConfigModel{
			Host: types.StringValue(c.Host),
			Port: types.Int64Value(c.Port),
			// Not optionalInt64: these are Optional+Computed because the server
			// always echoes them, so 0 must stay 0 rather than collapsing to null.
			WarnDaysBeforeExpiry:      types.Int64Value(c.WarnDaysBeforeExpiry),
			CriticalDaysBeforeExpiry:  types.Int64Value(c.CriticalDaysBeforeExpiry),
			InsecureSkipVerify:        types.BoolValue(c.InsecureSkipVerify),
			CheckCertificateAuthority: types.BoolValue(c.CheckCertificateAuthority),
		}
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
	if c := m.RuleConfig.Ping; c != nil {
		model.RuleConfig.Ping = &clientmodels.SyntheticMonitorPingConfig{
			Host:       c.Host.ValueString(),
			Count:      c.Count.ValueInt64(),
			IntervalMs: c.IntervalMs.ValueInt64(),
		}
	}
	if c := m.RuleConfig.DNS; c != nil {
		cfg := &clientmodels.SyntheticMonitorDNSConfig{
			Domain:           c.Domain.ValueString(),
			RecordType:       c.RecordType.ValueString(),
			Nameserver:       c.Nameserver.ValueString(),
			ExpectResolution: c.ExpectResolution.ValueBool(),
		}
		if len(c.ExpectedValues) > 0 {
			cfg.ExpectedValues = tfListToStrings(c.ExpectedValues)
		}
		model.RuleConfig.DNS = cfg
	}
	if c := m.RuleConfig.TCP; c != nil {
		model.RuleConfig.TCP = &clientmodels.SyntheticMonitorTCPConfig{
			Host: c.Host.ValueString(),
			Port: c.Port.ValueInt64(),
		}
	}
	if c := m.RuleConfig.Traceroute; c != nil {
		model.RuleConfig.Traceroute = &clientmodels.SyntheticMonitorTracerouteConfig{
			Host:            c.Host.ValueString(),
			MaxHops:         c.MaxHops.ValueInt64(),
			TimeoutPerHopMs: c.TimeoutPerHopMs.ValueInt64(),
		}
	}
	if c := m.RuleConfig.SSL; c != nil {
		model.RuleConfig.SSL = &clientmodels.SyntheticMonitorSSLConfig{
			Host:                      c.Host.ValueString(),
			Port:                      c.Port.ValueInt64(),
			WarnDaysBeforeExpiry:      c.WarnDaysBeforeExpiry.ValueInt64(),
			CriticalDaysBeforeExpiry:  c.CriticalDaysBeforeExpiry.ValueInt64(),
			InsecureSkipVerify:        c.InsecureSkipVerify.ValueBool(),
			CheckCertificateAuthority: c.CheckCertificateAuthority.ValueBool(),
		}
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
		Body:               optionalString(c.Body),
		ExpectedBody:       optionalString(c.ExpectedBody),
		MaxResponseTimeMs:  optionalInt64(c.MaxResponseTimeMs),
		BearerToken:        optionalString(c.BearerToken),
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
	if len(c.ExpectedHeaders) > 0 {
		cfg.ExpectedHeaders = c.ExpectedHeaders
	}
	if c.BasicAuth != nil {
		cfg.BasicAuth = &basicAuthModel{
			Username: types.StringValue(c.BasicAuth.Username),
			Password: types.StringValue(c.BasicAuth.Password),
		}
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

// optionalInt64 maps a zero value to null. The server stores these fields as
// sent and only substitutes its own defaults when the check runs, so a zero
// read back means "unset" rather than a server-assigned value.
func optionalInt64(v int64) types.Int64 {
	if v == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(v)
}

// optionalString maps an empty value to null, so omitted attributes round-trip
// as null instead of "".
func optionalString(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
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
