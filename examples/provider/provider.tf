terraform {
  required_providers {
    oodle = {
      source = "registry.terraform.io/hashicorp/oodle"
    }
  }
}

provider "oodle" {}

resource "oodle_monitor" "test1" {
  name = "terraform_test"
  promql_query = "sum(rate(oober_food_delivery_revenue_usd[3m]))"
  conditions = {
    critical = {
      value = 1240000
      operation = ">"
      for = "3m"
    }
  }
  notification_policy_id = "01918078-4424-762b-aaf9-ef33fc94fd51"
}
