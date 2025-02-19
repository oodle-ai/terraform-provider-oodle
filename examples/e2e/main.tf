terraform {
  required_providers {
    oodle = {
      source = "registry.terraform.io/oodle-ai/oodle"
    }
  }
}

provider "oodle" {}

# Notifiers for different teams/scenarios
resource "oodle_notifier" "platform_opsgenie" {
  name = "tf_platform_team_opsgenie"
  type = "opsgenie"
  opsgenie_config = {
    api_key       = "platform_team_key"
    send_resolved = true
  }
}

resource "oodle_notifier" "critical_slack" {
  name = "tf_critical_alerts_slack"
  type = "slack"
  slack_config = {
    api_url       = "https://hooks.slack.com/services/xxx/yyy/zzz"
    channel       = "#critical-alerts"
    send_resolved = true
  }
}

resource "oodle_notifier" "general_slack" {
  name = "tf_general_alerts_slack"
  type = "slack"
  slack_config = {
    api_url       = "https://hooks.slack.com/services/xxx/yyy/zzz"
    channel       = "#alerts"
    send_resolved = true
  }
}

resource "oodle_notifier" "general_googlechat" {
  name = "tf_general_alerts_googlechat"
  type = "googlechat"
  googlechat_config = {
    url           = "https://chat.googleapis.com/v1/spaces/XXXXXX/messages?key=YYYYYY&token=ZZZZZ"
    threading     = false
    send_resolved = true
  }
}

# Notification policies for different scenarios
resource "oodle_notification_policy" "platform_team" {
  name = "tf_platform_team_policy"
  notifiers = {
    critical = [oodle_notifier.platform_opsgenie.id]
    warning  = [oodle_notifier.general_slack.id]
  }
}

resource "oodle_notification_policy" "critical_services" {
  name = "tf_critical_services_policy"
  notifiers = {
    critical = [oodle_notifier.platform_opsgenie.id, oodle_notifier.critical_slack.id]
    warning  = [oodle_notifier.critical_slack.id]
  }
}

resource "oodle_notification_policy" "default" {
  name = "tf_default_policy"
  notifiers = {
    critical = [oodle_notifier.general_slack.id]
    warning  = [oodle_notifier.general_slack.id]
  }
}

# Monitor with label-based routing
resource "oodle_monitor" "service_monitor" {
  name         = "tf_service_health_monitor"
  promql_query = "sum(rate(service_errors_total[5m])) by (service, region, team) / sum(rate(service_requests_total[5m])) by (service, region, team) > 0.01"
  interval     = "1m"

  conditions = {
    warning = {
      value     = 0.01 # 1% error rate
      operation = ">"
      for       = "5m"
    }
    critical = {
      value     = 0.05 # 5% error rate
      operation = ">"
      for       = "3m"
    }
  }

  # Labels that will be attached to all alerts from this monitor
  labels = {
    monitor_type = "service_health"
    severity     = "high"
  }

  # Annotations provide additional context in notifications
  annotations = {
    summary     = "High error rate detected"
    description = "Service error rate has exceeded threshold"
    runbook_url = "https://wiki.example.com/runbooks/service-errors"
  }

  # Default policy for alerts that don't match any matchers
  notification_policy_id = oodle_notification_policy.default.id

  # Route alerts to different policies based on labels
  label_matcher_notification_policies = [
    {
      # Critical services get highest priority routing
      matchers = [
        {
          type  = "=~"
          name  = "service"
          value = "(auth|payment|core)-.*" # Regex to match critical services
        }
      ]
      notification_policy_id = oodle_notification_policy.critical_services.id
    },
    {
      # Platform team gets their own routing
      matchers = [
        {
          type  = "="
          name  = "team"
          value = "platform"
        }
      ]
      notification_policy_id = oodle_notification_policy.platform_team.id
    }
  ]
}
