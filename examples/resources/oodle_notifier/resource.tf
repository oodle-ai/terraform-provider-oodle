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
