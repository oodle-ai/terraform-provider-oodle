package syntheticmonitor

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/rubrikinc/testwell/assert"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestSyntheticMonitorModel(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.SyntheticMonitor{
		ID:       "test-id-123",
		Name:     "Test HTTP Monitor",
		Enabled:  true,
		RuleType: "http",
		RuleConfig: clientmodels.SyntheticMonitorRuleConfig{
			HTTP: &clientmodels.SyntheticMonitorHTTPConfig{
				URL:    "https://example.com",
				Method: "GET",
				Headers: map[string]string{
					"X-Custom": "header-value",
				},
				Body:                "request body",
				ExpectedStatusCodes: []string{"200", "201"},
				FollowRedirects:     true,
				InsecureSkipVerify:  false,
			},
		},
		Interval: "30s",
		Timeout:  "5s",
	}

	resourceModel := &syntheticMonitorResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.SyntheticMonitor{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}

func TestSyntheticMonitorModelMinimal(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.SyntheticMonitor{
		ID:       "test-id-456",
		Name:     "Simple Monitor",
		Enabled:  true,
		RuleType: "http",
		RuleConfig: clientmodels.SyntheticMonitorRuleConfig{
			HTTP: &clientmodels.SyntheticMonitorHTTPConfig{
				URL:                 "https://example.com",
				Method:              "GET",
				ExpectedStatusCodes: []string{"2XX"},
				FollowRedirects:     true,
				InsecureSkipVerify:  false,
			},
		},
		Interval: "1m",
		Timeout:  "10s",
	}

	resourceModel := &syntheticMonitorResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.SyntheticMonitor{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}

func TestSyntheticMonitorModelRichHTTP(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.SyntheticMonitor{
		ID:       "test-id-rich",
		Name:     "Rich HTTP Monitor",
		Enabled:  true,
		RuleType: "http",
		RuleConfig: clientmodels.SyntheticMonitorRuleConfig{
			HTTP: &clientmodels.SyntheticMonitorHTTPConfig{
				URL:                 "https://api.example.com/health",
				Method:              "GET",
				ExpectedStatusCodes: []string{"200"},
				ExcludedStatusCodes: []string{"5XX"},
				ExpectedBody:        "\"status\":\"ok\"",
				MaxResponseTimeMs:   800,
				ExpectedHeaders: map[string]string{
					"Content-Type": "application/json",
				},
				FollowRedirects:    true,
				InsecureSkipVerify: false,
				BasicAuth: &clientmodels.SyntheticMonitorBasicAuth{
					Username: "svc",
					Password: "secret",
				},
				BearerToken: "static-token",
			},
		},
		Interval: "1m",
		Timeout:  "10s",
	}

	resourceModel := &syntheticMonitorResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.SyntheticMonitor{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}

func TestSyntheticMonitorModelMultistep(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.SyntheticMonitor{
		ID:       "test-id-multistep",
		Name:     "Auth + Protected API",
		Enabled:  true,
		RuleType: "multistep",
		RuleConfig: clientmodels.SyntheticMonitorRuleConfig{
			Multistep: &clientmodels.SyntheticMonitorMultistepConfig{
				Steps: []clientmodels.SyntheticMonitorStep{
					{
						Name: "Get Token",
						Request: clientmodels.SyntheticMonitorHTTPConfig{
							URL:    "https://api.example.com/auth/token",
							Method: "POST",
							Headers: map[string]string{
								"Content-Type": "application/json",
							},
							Body:                "{\"client_id\":\"abc\"}",
							ExpectedStatusCodes: []string{"2XX"},
							FollowRedirects:     true,
							InsecureSkipVerify:  false,
						},
						Extract: []clientmodels.SyntheticMonitorExtractRule{
							{
								Name:   "ACCESS_TOKEN",
								Source: "body",
								Parser: "jsonpath",
								Query:  "$.access_token",
								Secret: true,
							},
							{
								Name:   "USER_ID",
								Source: "body",
								Parser: "jsonpath",
								Query:  "$.user.id",
							},
						},
					},
					{
						Name: "Get User Profile",
						Request: clientmodels.SyntheticMonitorHTTPConfig{
							URL:                "https://api.example.com/users/{{USER_ID}}",
							Method:             "GET",
							BearerToken:        "{{ACCESS_TOKEN}}",
							FollowRedirects:    false,
							InsecureSkipVerify: false,
						},
						ContinueOnFailure: true,
						ExitOnSuccess:     true,
					},
				},
			},
		},
		Interval: "5m",
		Timeout:  "30s",
	}

	resourceModel := &syntheticMonitorResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.SyntheticMonitor{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}

func TestSyntheticMonitorModelDisabled(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.SyntheticMonitor{
		ID:       "test-id-789",
		Name:     "Disabled Monitor",
		Enabled:  false,
		RuleType: "http",
		RuleConfig: clientmodels.SyntheticMonitorRuleConfig{
			HTTP: &clientmodels.SyntheticMonitorHTTPConfig{
				URL:                 "https://staging.example.com/health",
				Method:              "POST",
				Body:                "{\"check\": true}",
				ExpectedStatusCodes: []string{"200"},
				FollowRedirects:     false,
				InsecureSkipVerify:  true,
			},
		},
		Interval: "60s",
		Timeout:  "15s",
	}

	resourceModel := &syntheticMonitorResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	newClientModel := &clientmodels.SyntheticMonitor{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

	assert.DeepEqual(t, clientModel, newClientModel)
}

// TestSyntheticMonitorModelRuleTypes round-trips each non-HTTP rule type
// through the Terraform model and back. Each type is covered twice: once with
// every optional field populated, and once with only the required fields, which
// exercises the null-vs-zero handling in the conversion helpers.
func TestSyntheticMonitorModelRuleTypes(t *testing.T) {
	tests := []struct {
		name       string
		ruleType   string
		ruleConfig clientmodels.SyntheticMonitorRuleConfig
	}{
		{
			name:     "ping full",
			ruleType: "ping",
			ruleConfig: clientmodels.SyntheticMonitorRuleConfig{
				Ping: &clientmodels.SyntheticMonitorPingConfig{
					Host:       "example.com",
					Count:      5,
					IntervalMs: 500,
				},
			},
		},
		{
			name:     "ping minimal",
			ruleType: "ping",
			ruleConfig: clientmodels.SyntheticMonitorRuleConfig{
				Ping: &clientmodels.SyntheticMonitorPingConfig{
					Host: "example.com",
				},
			},
		},
		{
			name:     "dns full",
			ruleType: "dns",
			ruleConfig: clientmodels.SyntheticMonitorRuleConfig{
				DNS: &clientmodels.SyntheticMonitorDNSConfig{
					Domain:           "example.com",
					RecordType:       "MX",
					ExpectedValues:   []string{"mail.example.com (priority: 10)"},
					Nameserver:       "8.8.8.8:53",
					ExpectResolution: true,
				},
			},
		},
		{
			name:     "dns minimal",
			ruleType: "dns",
			ruleConfig: clientmodels.SyntheticMonitorRuleConfig{
				DNS: &clientmodels.SyntheticMonitorDNSConfig{
					Domain: "example.com",
				},
			},
		},
		{
			name:     "tcp",
			ruleType: "tcp",
			ruleConfig: clientmodels.SyntheticMonitorRuleConfig{
				TCP: &clientmodels.SyntheticMonitorTCPConfig{
					Host: "example.com",
					Port: 5432,
				},
			},
		},
		{
			name:     "traceroute full",
			ruleType: "traceroute",
			ruleConfig: clientmodels.SyntheticMonitorRuleConfig{
				Traceroute: &clientmodels.SyntheticMonitorTracerouteConfig{
					Host:            "example.com",
					MaxHops:         15,
					TimeoutPerHopMs: 2000,
				},
			},
		},
		{
			name:     "traceroute minimal",
			ruleType: "traceroute",
			ruleConfig: clientmodels.SyntheticMonitorRuleConfig{
				Traceroute: &clientmodels.SyntheticMonitorTracerouteConfig{
					Host: "example.com",
				},
			},
		},
		{
			name:     "ssl full",
			ruleType: "ssl",
			ruleConfig: clientmodels.SyntheticMonitorRuleConfig{
				SSL: &clientmodels.SyntheticMonitorSSLConfig{
					Host:                      "example.com",
					Port:                      443,
					WarnDaysBeforeExpiry:      30,
					CriticalDaysBeforeExpiry:  7,
					InsecureSkipVerify:        false,
					CheckCertificateAuthority: true,
				},
			},
		},
		{
			name:     "ssl minimal",
			ruleType: "ssl",
			ruleConfig: clientmodels.SyntheticMonitorRuleConfig{
				SSL: &clientmodels.SyntheticMonitorSSLConfig{
					Host: "example.com",
					Port: 443,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			clientModel := &clientmodels.SyntheticMonitor{
				ID:         "test-id-" + tt.ruleType,
				Name:       "Test " + tt.ruleType + " Monitor",
				Enabled:    true,
				RuleType:   tt.ruleType,
				RuleConfig: tt.ruleConfig,
				Interval:   "1m",
				Timeout:    "10s",
			}

			resourceModel := &syntheticMonitorResourceModel{}
			diags := &diag.Diagnostics{}
			resourceModel.FromClientModel(ctx, clientModel, diags)
			assert.False(t, diags.HasError())

			newClientModel := &clientmodels.SyntheticMonitor{}
			assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))

			assert.DeepEqual(t, clientModel, newClientModel)
		})
	}
}

// TestSyntheticMonitorModelOptionalsAreNull guards the reason the optional
// scalar attributes are not Computed: the server stores them as sent and only
// substitutes defaults when the check runs. A zero read back therefore means
// "unset" and must round-trip as null, not as a concrete 0.
func TestSyntheticMonitorModelOptionalsAreNull(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.SyntheticMonitor{
		ID:       "test-id-nulls",
		Name:     "Minimal Ping",
		Enabled:  true,
		RuleType: "ping",
		RuleConfig: clientmodels.SyntheticMonitorRuleConfig{
			Ping: &clientmodels.SyntheticMonitorPingConfig{Host: "example.com"},
		},
		Interval: "1m",
		Timeout:  "10s",
	}

	resourceModel := &syntheticMonitorResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	assert.True(t, resourceModel.RuleConfig.Ping.Count.IsNull())
	assert.True(t, resourceModel.RuleConfig.Ping.IntervalMs.IsNull())
}

// TestSyntheticMonitorSSLExpiryDaysRoundTripZero guards the Optional+Computed
// choice on the SSL expiry thresholds. The server echoes these fields whether or
// not they are set, so an explicit 0 (a reasonable way to spell "disabled") must
// survive the round trip as 0. Collapsing it to null would make every plan show
// a diff that apply can never settle.
func TestSyntheticMonitorSSLExpiryDaysRoundTripZero(t *testing.T) {
	ctx := context.Background()
	clientModel := &clientmodels.SyntheticMonitor{
		ID:       "test-id-ssl-zero",
		Name:     "SSL No Thresholds",
		Enabled:  true,
		RuleType: "ssl",
		RuleConfig: clientmodels.SyntheticMonitorRuleConfig{
			SSL: &clientmodels.SyntheticMonitorSSLConfig{
				Host: "example.com",
				Port: 443,
			},
		},
		Interval: "1h",
		Timeout:  "10s",
	}

	resourceModel := &syntheticMonitorResourceModel{}
	diags := &diag.Diagnostics{}
	resourceModel.FromClientModel(ctx, clientModel, diags)
	assert.False(t, diags.HasError())

	ssl := resourceModel.RuleConfig.SSL
	assert.False(t, ssl.WarnDaysBeforeExpiry.IsNull())
	assert.False(t, ssl.CriticalDaysBeforeExpiry.IsNull())
	assert.Equal(t, int64(0), ssl.WarnDaysBeforeExpiry.ValueInt64())
	assert.Equal(t, int64(0), ssl.CriticalDaysBeforeExpiry.ValueInt64())

	newClientModel := &clientmodels.SyntheticMonitor{}
	assert.Nil(t, resourceModel.ToClientModel(ctx, newClientModel))
	assert.DeepEqual(t, clientModel, newClientModel)
}
