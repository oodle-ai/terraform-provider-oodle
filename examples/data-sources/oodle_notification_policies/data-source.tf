data "oodle_notification_policies" "all" {}

output "notification_policies" {
  value = data.oodle_notification_policies.all.notification_policies
}
