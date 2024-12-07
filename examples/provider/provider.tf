terraform {
  required_providers {
    oodle = {
      source = "registry.terraform.io/hashicorp/oodle"
    }
  }
}

provider "oodle" {}

resource "oodle_notifier" "notifier_test1" {
  name = "terraform_test_notifier"
  type = "pagerduty"
  pagerduty_config = {
    service_key = "foo"
    send_resolved = true
  }
}

resource "oodle_notification_policy" "test1" {
  name = "terraform_test_policy"
  notifiers = {
    critical = [oodle_notifier.notifier_test1.id]
  }
}

resource "oodle_monitor" "test1" {
  name = "terraform_test"
  promql_query = "sum(rate(oober_food_delivery_revenue_usd[3m]))"
  conditions = {
    critical = {
      value = 1210000
      operation = ">"
      for = "3m"
    }
  }
  notification_policy_id = oodle_notification_policy.test1.id
}
