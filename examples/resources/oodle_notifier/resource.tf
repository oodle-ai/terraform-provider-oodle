# Email notifier for critical alerts
resource "oodle_notifier" "critical_email" {
  name = "critical_alerts_email"
  type = "email"
  email_config = {
    to            = "alerts@company.com"
    send_resolved = true
  }
}

# Opsgenie notifier for critical alerts from platform team
resource "oodle_notifier" "platform_opsgenie" {
  name = "platform_team_opsgenie"
  type = "opsgenie"
  opsgenie_config = {
    api_key       = "platform_team_key"
    send_resolved = true
  }
}

# Slack notifier for critical service alerts
resource "oodle_notifier" "critical_slack" {
  name = "critical_alerts_slack"
  type = "slack"
  slack_config = {
    api_url       = "https://hooks.slack.com/services/xxx/yyy/zzz"
    channel       = "#critical-alerts"
    send_resolved = true
  }
}

# General Slack notifier for all other alerts
resource "oodle_notifier" "general_slack" {
  name = "general_alerts_slack"
  type = "slack"
  slack_config = {
    api_url       = "https://hooks.slack.com/services/xxx/yyy/zzz"
    channel       = "#alerts"
    send_resolved = true
  }
}

# Microsoft Teams notifier
resource "oodle_notifier" "team_msteams" {
  name = "platform_team_msteams"
  type = "msteamsv2"
  msteamsv2_config = {
    webhook_url   = "https://company.webhook.office.com/webhookb2/xxx/IncomingWebhook/yyy/zzz"
    send_resolved = true
  }
}

# Rootly notifier for incident coordination and on-call paging.
# Create an Alertmanager alert source in Rootly and paste its bearer token here;
# `url` defaults to the Rootly Alertmanager webhook endpoint.
resource "oodle_notifier" "incident_rootly" {
  name = "incidents_rootly"
  type = "rootly"
  rootly_config = {
    bearer_token  = "rootly_alert_source_bearer_token"
    send_resolved = true
  }
}

# Webhook notifier posting Oodle's default alert payload.
resource "oodle_notifier" "generic_webhook" {
  name = "generic_webhook"
  type = "webhook"
  webhook_config = {
    url           = "https://example.com/hooks/oodle"
    send_resolved = true
  }
}

# Webhook notifier with a fully custom payload. Every string is a Go template
# rendered against the alert data; `toJson` embeds structured values.
resource "oodle_notifier" "custom_payload_webhook" {
  name = "custom_payload_webhook"
  type = "webhook"
  webhook_config = {
    url           = "https://example.com/hooks/oodle"
    send_resolved = true
    payload = jsonencode({
      text   = "{{ .CommonLabels.alertname }} is {{ .Status }}"
      status = "{{ .Status }}"
      labels = "{{ .CommonLabels | toJson }}"
      alerts = [
        {
          summary = "{{ .CommonAnnotations.summary }}"
        }
      ]
    })
  }
}
