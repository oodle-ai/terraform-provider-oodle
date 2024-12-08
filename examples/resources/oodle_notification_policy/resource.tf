resource "oodle_notification_policy" "test1" {
  name = "terraform_test_policy"
  notifiers = {
    critical = ["oodle_notifier.notifier_test1.id"]
  }
}
