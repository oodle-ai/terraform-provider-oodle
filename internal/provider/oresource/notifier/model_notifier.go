// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifier

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prometheus/alertmanager/config"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
	"terraform-provider-oodle/internal/oodlehttp/clientmodels/oprom"
	"terraform-provider-oodle/internal/resourceutils"
)

type notifierResourceModel struct {
	ID              types.String          `tfsdk:"id"`
	Name            types.String          `tfsdk:"name"`
	Type            types.String          `tfsdk:"type"`
	PagerdutyConfig *pagerdutyConfigModel `tfsdk:"pagerduty_config"`
	SlackConfig     *slackConfigModel     `tfsdk:"slack_config"`
	OpsGenieConfig  *opsgenieConfigModel  `tfsdk:"opsgenie_config"`
	WebhookConfig   *webhookConfigModel   `tfsdk:"webhook_config"`
}

var _ resourceutils.ResourceModel[*clientmodels.Notifier] = (*notifierResourceModel)(nil)

func (n *notifierResourceModel) GetID() types.String {
	return n.ID
}

func (n *notifierResourceModel) SetID(id types.String) {
	n.ID = id
}

func (n *notifierResourceModel) FromClientModel(
	model *clientmodels.Notifier,
	diagnosticsOut *diag.Diagnostics,
) {
	n.ID = types.StringValue(model.ID.UUID.String())
	n.Name = types.StringValue(model.Name)
	n.Type = types.StringValue(model.Type.String())
	if len(n.Type.ValueString()) == 0 {
		diagnosticsOut.AddError("Unknown type", fmt.Sprintf("Unknown notifier type %v", model.Type))
	}

	switch model.Type {
	case clientmodels.NotifierConfigPagerduty:
		if model.PagerdutyConfig == nil {
			diagnosticsOut.AddError("Missing PagerDuty config", "PagerDuty config is required for PagerDuty notifier")
			return
		}

		n.PagerdutyConfig = &pagerdutyConfigModel{}
		n.PagerdutyConfig.SendResolved = types.BoolValue(model.PagerdutyConfig.SendResolved())
		n.PagerdutyConfig.ServiceKey = types.StringValue(model.PagerdutyConfig.ServiceKey)
	case clientmodels.NotifierConfigSlack:
		if model.SlackConfig == nil {
			diagnosticsOut.AddError("Missing Slack config", "Slack config is required for Slack notifier")
			return
		}

		n.SlackConfig = &slackConfigModel{}
		n.SlackConfig.SendResolved = types.BoolValue(model.SlackConfig.SendResolved())
		n.SlackConfig.APIURL = types.StringValue(model.SlackConfig.APIURL)
		n.SlackConfig.Channel = types.StringValue(model.SlackConfig.Channel)
		n.SlackConfig.TitleLink = types.StringValue(model.SlackConfig.TitleLink)
		n.SlackConfig.Text = types.StringValue(model.SlackConfig.Text)
	case clientmodels.NotifierConfigOpsGenie:
		if model.OpsGenieConfig == nil {
			diagnosticsOut.AddError("Missing OpsGenie config", "OpsGenie config is required for OpsGenie notifier")
			return
		}

		n.OpsGenieConfig = &opsgenieConfigModel{}
		n.OpsGenieConfig.SendResolved = types.BoolValue(model.OpsGenieConfig.SendResolved())
		n.OpsGenieConfig.APIKey = types.StringValue(model.OpsGenieConfig.APIKey)
	case clientmodels.NotifierConfigWebhook:
		if model.WebhookConfig == nil {
			diagnosticsOut.AddError("Missing Webhook config", "Webhook config is required for Webhook notifier")
			return
		}

		n.WebhookConfig = &webhookConfigModel{}
		n.WebhookConfig.SendResolved = types.BoolValue(model.WebhookConfig.SendResolved())
		n.WebhookConfig.URL = types.StringValue(model.WebhookConfig.URL)
	default:
		diagnosticsOut.AddError("Unknown type", fmt.Sprintf("Unknown notifier type %v", model.Type))
		return
	}
}

func (n *notifierResourceModel) ToClientModel(
	model *clientmodels.Notifier,
) error {
	var err error
	if !n.ID.IsNull() && !n.ID.IsUnknown() {
		model.ID.UUID, err = uuid.Parse(n.ID.ValueString())
		if err != nil {
			return fmt.Errorf("failed to parse ID UUID %v: %v", n.ID.ValueString(), err)
		}
	}

	model.Name = n.Name.ValueString()
	model.Type, err = clientmodels.GetNotifierType(n.Type.ValueString())
	if err != nil {
		return fmt.Errorf("failed to parse notifier type %v: %v", n.Type.ValueString(), err)
	}

	switch model.Type {
	case clientmodels.NotifierConfigPagerduty:
		if n.PagerdutyConfig == nil {
			return fmt.Errorf("missing PagerDuty config")
		}

		model.PagerdutyConfig = &oprom.PagerdutyConfig{
			ServiceKey: n.PagerdutyConfig.ServiceKey.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.PagerdutyConfig.SendResolved.ValueBool(),
			},
		}
	case clientmodels.NotifierConfigSlack:
		if n.SlackConfig == nil {
			return fmt.Errorf("missing Slack config")
		}

		model.SlackConfig = &oprom.SlackConfig{
			APIURL:    n.SlackConfig.APIURL.ValueString(),
			Channel:   n.SlackConfig.Channel.ValueString(),
			TitleLink: n.SlackConfig.TitleLink.ValueString(),
			Text:      n.SlackConfig.Text.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.SlackConfig.SendResolved.ValueBool(),
			},
		}
	case clientmodels.NotifierConfigOpsGenie:
		if n.OpsGenieConfig == nil {
			return fmt.Errorf("missing OpsGenie config")
		}

		model.OpsGenieConfig = &oprom.OpsGenieConfig{
			APIKey: n.OpsGenieConfig.APIKey.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.OpsGenieConfig.SendResolved.ValueBool(),
			},
		}
	case clientmodels.NotifierConfigWebhook:
		if n.WebhookConfig == nil {
			return fmt.Errorf("missing Webhook config")
		}

		model.WebhookConfig = &oprom.WebhookConfig{
			URL: n.WebhookConfig.URL.ValueString(),
			NotifierConfig: config.NotifierConfig{
				VSendResolved: n.WebhookConfig.SendResolved.ValueBool(),
			},
		}
	default:
		return fmt.Errorf("unknown notifier type %v", model.Type)
	}

	return nil
}
