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
    warn     = [oodle_notifier.general_slack.id]
  }
}

resource "oodle_notification_policy" "critical_services" {
  name = "tf_critical_services_policy"
  notifiers = {
    critical = [oodle_notifier.platform_opsgenie.id, oodle_notifier.critical_slack.id]
    warn     = [oodle_notifier.critical_slack.id]
  }
}

resource "oodle_notification_policy" "default" {
  name = "tf_default_policy"
  notifiers = {
    critical = [oodle_notifier.general_slack.id]
    warn     = [oodle_notifier.general_slack.id]
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
      for       = "0s"
    }
    no_data = {
      "for" = "5m"
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
    },
    {
      notifiers = {
        any = [oodle_notifier.general_slack.id]
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

resource "oodle_grafana_folder" "my_folder" {
  title = "Terraform - Test Folder"
}

# Example of a dashboard with a commit message for version history
resource "oodle_grafana_dashboard" "service_dashboard" {
  folder    = oodle_grafana_folder.my_folder.uid
  message   = "Initial dashboard creation"
  overwrite = true
  config_json = jsonencode({
    "title" : "Terraform - Service Health",
    "uid" : "service-health-dashboard",
    "schemaVersion" : 39,
    "time" : {
      "from" : "now-6h",
      "to" : "now"
    },
    "panels" : [
      {
        "id" : 1,
        "type" : "gauge",
        "title" : "Uptime",
        "gridPos" : {
          "h" : 8,
          "w" : 8,
          "x" : 0,
          "y" : 0
        },
        "targets" : [
          {
            "expr" : "avg(up{})",
            "refId" : "A"
          }
        ]
      }
    ]
  })
}

# --- GenAI ---

resource "oodle_genai_llm_connection" "tf_genai" {
  name          = "tf_genai_connection"
  llm_provider  = "openai"
  api_key       = "sk-placeholder-not-a-real-key"
  default_model = "gpt-4o-mini"
  custom_models = ["gpt-4o", "gpt-4o-mini"]

  default_params = jsonencode({
    temperature = 0
  })
}

resource "oodle_genai_eval_template" "tf_genai" {
  name = "tf_genai_template"
  type = "llm"

  prompt = <<-EOT
    You are grading a support reply.

    Question: {{question}}
    Answer: {{answer}}

    Score 1 if the answer resolves the question, otherwise 0.
  EOT

  vars = ["question", "answer"]

  output_schema = jsonencode({
    score     = "0 or 1"
    reasoning = "one sentence explaining the score"
  })
}

resource "oodle_genai_evaluator" "tf_genai" {
  name              = "tf_genai_evaluator"
  eval_template_id  = oodle_genai_eval_template.tf_genai.id
  llm_connection_id = oodle_genai_llm_connection.tf_genai.id

  enabled                  = false
  target_type              = "trace"
  sampling_rate            = 0.1
  max_invocations_per_hour = 100

  variable_mapping = jsonencode([
    {
      templateVariable = "question"
      langfuseObject   = "trace"
      selectedColumnId = "input"
    },
    {
      templateVariable = "answer"
      langfuseObject   = "trace"
      selectedColumnId = "output"
    },
  ])
}

resource "oodle_genai_dataset" "tf_genai" {
  name        = "tf_genai_dataset"
  description = "Created by the Terraform end-to-end example."

  metadata = jsonencode({
    owner = "terraform"
  })
}

resource "oodle_genai_dataset_item" "tf_genai" {
  dataset_name = oodle_genai_dataset.tf_genai.name

  input           = jsonencode({ question = "How do I reset my password?" })
  expected_output = jsonencode({ answer = "Use the reset link on the login page." })
}

resource "oodle_genai_prompt" "tf_genai" {
  name   = "tf_genai_prompt"
  type   = "text"
  prompt = "Answer the customer question concisely: {{question}}"

  labels         = ["production"]
  tags           = ["terraform"]
  commit_message = "Created by the Terraform end-to-end example."

  config = jsonencode({
    model = "gpt-4o-mini"
  })
}

# --- Muting rules ---

resource "oodle_muting_rule" "tf_muting_window" {
  comment = "Terraform end-to-end example: muted for a scheduled migration."

  matchers = [
    {
      type  = "="
      name  = "_oodle_monitor_id"
      value = oodle_monitor.service_monitor.id
    },
  ]

  starts_at = "2027-01-01T22:00:00Z"
  ends_at   = "2027-01-02T04:00:00Z"
}

resource "oodle_muting_rule" "tf_muting_staging" {
  comment = "Terraform end-to-end example: staging alerts are not actionable."

  matchers = [
    {
      type  = "="
      name  = "_oodle_monitor_id"
      value = oodle_monitor.service_monitor.id
    },
    {
      type  = "="
      name  = "environment"
      value = "staging"
    },
  ]
}
