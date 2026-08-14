# Mute a monitor during a planned migration. A muting rule is always
# scoped to a monitor, so an equality matcher on _oodle_monitor_id is
# required.
resource "oodle_muting_rule" "migration_window" {
  comment = "Muted for the scheduled database migration; see CHANGE-4821."

  matchers = [
    {
      type  = "="
      name  = "_oodle_monitor_id"
      value = oodle_monitor.service_monitor.id
    },
  ]

  starts_at = "2026-09-01T22:00:00Z"
  ends_at   = "2026-09-02T04:00:00Z"
}

# Narrow the mute to part of a monitor's alerts by adding matchers.
resource "oodle_muting_rule" "staging_only" {
  comment = "Staging alerts are not actionable."

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

  # No ends_at, so the mute is effective until the rule is removed.
}

# A recurring rule mutes on a schedule instead of over a fixed window.
# Recurring rules are the only kind that store a name, and they take
# exactly one matcher today, so they mute the whole monitor.
resource "oodle_muting_rule" "overnight" {
  name = "Overnight maintenance window"

  matchers = [
    {
      type  = "="
      name  = "_oodle_monitor_id"
      value = oodle_monitor.service_monitor.id
    },
  ]

  schedule_ids = [var.overnight_schedule_id]
}
