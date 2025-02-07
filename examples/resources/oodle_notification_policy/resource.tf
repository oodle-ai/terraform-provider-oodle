# Platform team's notification policy - Opsgenie for critical, general Slack for warnings
resource "oodle_notification_policy" "platform_team" {
  name = "platform_team_policy"
  notifiers = {
    critical = [oodle_notifier.platform_opsgenie.id]
    warning  = [oodle_notifier.general_slack.id]
  }
}

# Critical services policy - both Opsgenie and dedicated Slack for critical, general Slack for warnings
resource "oodle_notification_policy" "critical_services" {
  name = "critical_services_policy"
  notifiers = {
    critical = [oodle_notifier.platform_opsgenie.id, oodle_notifier.critical_slack.id]
    warning  = [oodle_notifier.critical_slack.id]
  }
}

# Default notification policy - general Slack for all alerts
resource "oodle_notification_policy" "default" {
  name = "default_policy"
  notifiers = {
    critical = [oodle_notifier.general_slack.id]
    warning  = [oodle_notifier.general_slack.id]
  }
}
