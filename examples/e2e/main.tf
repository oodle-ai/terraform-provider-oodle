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

resource "oodle_notifier" "critical_email" {
  name = "tf_critical_alerts_email"
  type = "email"
  email_config = {
    to            = "test@example.com"
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
      value            = 0.05 # 5% error rate
      operation        = ">"
      for              = "3m"
      alert_on_no_data = true # Alert if no data is received
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

  grouping = {
    disabled = true
  }

  # Route alerts to different policies based on labels using the new notifications field
  notifications = [
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
      # Platform team gets their own routing with custom notifiers
      matchers = [
        {
          type  = "="
          name  = "team"
          value = "platform"
        }
      ]
      notifiers = {
        any = [oodle_notifier.platform_opsgenie.id, oodle_notifier.critical_slack.id]

      }
    },
    {
      # Development services get simple notifications to any severity
      matchers = [
        {
          type  = "="
          name  = "environment"
          value = "development"
        }
      ]
      notifiers = {
        warn     = [oodle_notifier.general_slack.id]
        critical = [oodle_notifier.general_slack.id]
      }
    }
  ]
}

resource "oodle_logmetrics" "coverage" {
  name = "tf_app_coverage"

  labels = [
    {
      name  = "environment"
      value = "prod"
    },
    {
      name = "container"
      value_extractor = {
        field = "container_name"
      }
    },
    {
      name = "service",
      value_extractor = {
        field     = "log"
        json_path = "service.id"
      }
    },
    {
      name = "step",
      value_extractor = {
        field = "message"
        regex = "step=(\\w+)"
      }
    }
  ]

  filter = {
    any = [{
      all = [{
        match = {
          field    = "level"
          operator = "is"
          value    = "error"
        },
        },
        {
          match = {
            field    = "container"
            operator = "matches regex"
            value    = "(checkout|payment)"
          },
        },
        {
          match = {
            field     = "log"
            operator  = "contains"
            json_path = "service.id"
            value     = "123"
          },
        },
        {
          match = {
            field    = "namespace"
            operator = "exists"
          }
        },
        {
          not = {
            match = {
              field    = "message"
              operator = "contains"
              value    = "test"
            }
          }
        }
      ],
      },
      {
        not = {
          match = {
            field    = "container"
            operator = "is"
            value    = "otel-demo"
          }
        }
    }]
  }

  metric_definitions = [
    {
      name = "oodle_logs_app_log_count"
      type = "log_count"
    },
    {
      name  = "oodle_logs_app_thread_count"
      type  = "gauge"
      field = "thread_count"
    },
    {
      name      = "oodle_logs_app_duration"
      type      = "histogram"
      field     = "log",
      json_path = "duration"
    },
    {
      name  = "oodle_logs_app_step_count"
      type  = "counter"
      field = "step",
      regex = "step=(\\w+)"
    }
  ]
}
