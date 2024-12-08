resource "oodle_notifier" "notifier_test1" {
  name = "terraform_test_notifier"
  type = "pagerduty"
  pagerduty_config = {
    service_key   = "foo"
    send_resolved = true
  }
}
