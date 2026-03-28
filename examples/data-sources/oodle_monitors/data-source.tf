data "oodle_monitors" "all" {}

output "monitors" {
  value = data.oodle_monitors.all.monitors
}
