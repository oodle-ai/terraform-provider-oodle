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
