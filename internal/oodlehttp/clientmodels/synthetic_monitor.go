package clientmodels

// SyntheticMonitorHTTPConfig represents the HTTP configuration for a synthetic monitor.
type SyntheticMonitorHTTPConfig struct {
	// URL is the URL to monitor.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`

	// Method is the HTTP method to use.
	Method string `json:"method,omitempty" yaml:"method,omitempty"`

	// Headers are optional HTTP headers to send with the request.
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Body is an optional request body.
	Body string `json:"body,omitempty" yaml:"body,omitempty"`

	// ExpectedStatusCodes is a list of expected HTTP status codes or patterns (e.g., "2XX").
	ExpectedStatusCodes []string `json:"expected_status_codes" yaml:"expected_status_codes"`

	// FollowRedirects indicates whether to follow HTTP redirects.
	FollowRedirects bool `json:"follow_redirects" yaml:"follow_redirects"`

	// InsecureSkipVerify indicates whether to skip TLS certificate verification.
	InsecureSkipVerify bool `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// SyntheticMonitorRuleConfig represents the rule configuration for a synthetic monitor.
type SyntheticMonitorRuleConfig struct {
	// HTTP is the HTTP rule configuration.
	HTTP *SyntheticMonitorHTTPConfig `json:"http,omitempty" yaml:"http,omitempty"`
}

// SyntheticMonitor represents a synthetic monitor definition.
type SyntheticMonitor struct {
	// ID is the unique identifier.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// Name is the human-readable name for the synthetic monitor.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Enabled indicates whether the synthetic monitor is active.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// RuleType is the type of the synthetic monitor rule (e.g., "http").
	RuleType string `json:"rule_type,omitempty" yaml:"rule_type,omitempty"`

	// RuleConfig is the configuration for the synthetic monitor rule.
	RuleConfig SyntheticMonitorRuleConfig `json:"rule_config" yaml:"rule_config"`

	// Interval is the interval between checks (e.g., "30s", "1m").
	Interval string `json:"interval,omitempty" yaml:"interval,omitempty"`

	// Timeout is the timeout for each check (e.g., "5s", "10s").
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// GetID returns the ID of the synthetic monitor.
func (s *SyntheticMonitor) GetID() string {
	return s.ID
}
