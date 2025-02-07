resource "oodle_monitor" "service_monitor" {
  name         = "service_health_monitor"
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
