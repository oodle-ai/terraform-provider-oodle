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
  notification_policy_id = "oodle_notification_policy.test1.id"
}
