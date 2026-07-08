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
