# Mute outside working hours, in the team's own timezone. The schedule
# is active whenever the current time falls inside a time interval, so
# the two intervals below cover the night either side of the workday.
resource "oodle_muting_schedule" "outside_business_hours" {
  name = "outside-business-hours"

  time_intervals = [
    {
      times = [
        { start_time = "00:00", end_time = "09:00" },
        { start_time = "17:00", end_time = "23:59" },
      ]
      weekdays = ["monday", "tuesday", "wednesday", "thursday", "friday"]
      location = "America/New_York"
    },
    {
      # Weekends in full: no times, so the whole day is covered.
      weekdays = ["saturday", "sunday"]
      location = "America/New_York"
    },
  ]
}

# The schedule mutes nothing on its own. A muting rule references it,
# which makes the rule recurring rather than one-off.
resource "oodle_muting_rule" "quiet_hours" {
  name = "Quiet hours"

  matchers = [
    {
      type  = "="
      name  = "_oodle_monitor_id"
      value = oodle_monitor.service_monitor.id
    },
  ]

  schedule_ids = [oodle_muting_schedule.outside_business_hours.id]
}

# A maintenance window that recurs on the first weekend of the month.
# Constraints within one interval are combined, so this is the first
# five days of the month, and only the Saturday among them.
resource "oodle_muting_schedule" "monthly_maintenance" {
  name = "monthly-maintenance"

  time_intervals = [
    {
      times         = [{ start_time = "02:00", end_time = "06:00" }]
      weekdays      = ["saturday"]
      days_of_month = ["1:5"]
      location      = "UTC"
    },
  ]
}

# A window limited to a stretch of the calendar: the end of every
# month, through the end of 2027 only.
resource "oodle_muting_schedule" "month_end_close" {
  name = "month-end-close"

  time_intervals = [
    {
      # -3:-1 is the last three days of whichever month it lands in.
      days_of_month = ["-3:-1"]
      months        = ["january:december"]
      years         = ["2026:2027"]
      location      = "Europe/London"
    },
  ]
}
