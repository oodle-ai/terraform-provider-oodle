terraform {
  required_providers {
    oodle = {
      source = "registry.terraform.io/oodle-ai/oodle"
    }
  }
}

# These can also be set as environment variables:
# export OODLE_DEPLOYMENT=https://us1.oodle.ai/
# export OODLE_INSTANCE="my-instance"
# export OODLE_API_KEY="my-api-key"
provider "oodle" {
  deployment_url = "https://us1.oodle.ai/"
  instance       = "my-instance"
  api_key        = "my-api-key"
}

# Example usage of notifier, notification policy and monitor.
# Refer to resource documentation on all configurable fields
# of these resources.
resource "oodle_notifier" "notifier_test1" {
  name = "terraform_test_notifier"
  type = "pagerduty"
  pagerduty_config = {
    service_key   = "foo"
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
  name         = "terraform_test"
  promql_query = "sum(rate(oober_food_delivery_revenue_usd[3m]))"
  conditions = {
    critical = {
      value     = 1210000
      operation = ">"
      for       = "3m"
    }
  }
  notification_policy_id = oodle_notification_policy.test1.id
}
