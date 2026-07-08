package clientmodels

// SyntheticMonitorBasicAuth represents HTTP basic authentication credentials.
type SyntheticMonitorBasicAuth struct {
	// Username is the basic auth username.
	Username string `json:"username" yaml:"username"`

	// Password is the basic auth password.
	Password string `json:"password" yaml:"password"`
}

// SyntheticMonitorHTTPConfig represents the HTTP configuration for a synthetic
// monitor. It is shared by single-step ("http") monitors and by each request
// in a multi-step monitor.
type SyntheticMonitorHTTPConfig struct {
	// URL is the URL to monitor. In multi-step monitors it may reference
	// variables extracted from earlier steps using {{VAR_NAME}} syntax.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`

	// Method is the HTTP method to use.
	Method string `json:"method,omitempty" yaml:"method,omitempty"`

	// Headers are optional HTTP headers to send with the request.
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Body is an optional request body.
	Body string `json:"body,omitempty" yaml:"body,omitempty"`

	// ExpectedStatusCodes is a list of expected HTTP status codes or patterns (e.g., "2XX").
	ExpectedStatusCodes []string `json:"expected_status_codes,omitempty" yaml:"expected_status_codes,omitempty"`

	// ExcludedStatusCodes is a list of status codes or patterns that cause the check to fail.
	ExcludedStatusCodes []string `json:"excluded_status_codes,omitempty" yaml:"excluded_status_codes,omitempty"`

	// ExpectedBody is an optional substring that must appear in the response body.
	ExpectedBody string `json:"expected_body,omitempty" yaml:"expected_body,omitempty"`

	// MaxResponseTimeMs fails the check if the response takes longer than this (milliseconds).
	MaxResponseTimeMs int64 `json:"max_response_time_ms,omitempty" yaml:"max_response_time_ms,omitempty"`

	// ExpectedHeaders is a map of response header names to expected values.
	ExpectedHeaders map[string]string `json:"expected_headers,omitempty" yaml:"expected_headers,omitempty"`

	// FollowRedirects indicates whether to follow HTTP redirects.
	FollowRedirects bool `json:"follow_redirects" yaml:"follow_redirects"`

	// InsecureSkipVerify indicates whether to skip TLS certificate verification.
	InsecureSkipVerify bool `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`

	// BasicAuth holds optional HTTP basic authentication credentials.
	BasicAuth *SyntheticMonitorBasicAuth `json:"basic_auth,omitempty" yaml:"basic_auth,omitempty"`

	// BearerToken is an optional bearer token sent as an Authorization header.
	BearerToken string `json:"bearer_token,omitempty" yaml:"bearer_token,omitempty"`
}

// SyntheticMonitorExtractRule pulls a value from a step's response into a named
// variable that later steps can interpolate with {{VAR_NAME}}.
type SyntheticMonitorExtractRule struct {
	// Name is the variable name (uppercase, letter-first, at least 3 characters).
	Name string `json:"name" yaml:"name"`

	// Source is where to read from: "body" or "header".
	Source string `json:"source" yaml:"source"`

	// Parser is how to extract the value: "jsonpath", "regex", or "header_value".
	Parser string `json:"parser" yaml:"parser"`

	// Query is the JSONPath expression, regex (first capture group), or header name.
	Query string `json:"query" yaml:"query"`

	// Secret marks the extracted value as redacted in results and logs.
	Secret bool `json:"secret,omitempty" yaml:"secret,omitempty"`
}

// SyntheticMonitorStep is a single request in a multi-step synthetic monitor.
type SyntheticMonitorStep struct {
	// Name is a human-readable label for the step.
	Name string `json:"name" yaml:"name"`

	// Request is the HTTP request configuration for this step.
	Request SyntheticMonitorHTTPConfig `json:"request" yaml:"request"`

	// Extract pulls values from this step's response into variables for later steps.
	Extract []SyntheticMonitorExtractRule `json:"extract,omitempty" yaml:"extract,omitempty"`

	// ContinueOnFailure lets the chain proceed even if this step fails.
	ContinueOnFailure bool `json:"continue_on_failure,omitempty" yaml:"continue_on_failure,omitempty"`

	// ExitOnSuccess ends the chain early (marking the monitor passed) when this step succeeds.
	ExitOnSuccess bool `json:"exit_on_success,omitempty" yaml:"exit_on_success,omitempty"`
}

// SyntheticMonitorMultistepConfig is the configuration for a multi-step monitor.
type SyntheticMonitorMultistepConfig struct {
	// Steps is the ordered list of requests to execute (1-20 steps).
	Steps []SyntheticMonitorStep `json:"steps" yaml:"steps"`
}

// SyntheticMonitorRuleConfig represents the rule configuration for a synthetic monitor.
type SyntheticMonitorRuleConfig struct {
	// HTTP is the HTTP rule configuration (rule_type "http").
	HTTP *SyntheticMonitorHTTPConfig `json:"http,omitempty" yaml:"http,omitempty"`

	// Multistep is the multi-step rule configuration (rule_type "multistep").
	Multistep *SyntheticMonitorMultistepConfig `json:"multistep,omitempty" yaml:"multistep,omitempty"`
}

// SyntheticMonitor represents a synthetic monitor definition.
type SyntheticMonitor struct {
	// ID is the unique identifier.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// Name is the human-readable name for the synthetic monitor.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Enabled indicates whether the synthetic monitor is active.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// RuleType is the type of the synthetic monitor rule (e.g., "http", "multistep").
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
