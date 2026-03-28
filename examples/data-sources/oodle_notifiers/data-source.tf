data "oodle_notifiers" "all" {}

output "notifiers" {
  value = data.oodle_notifiers.all.notifiers
}
