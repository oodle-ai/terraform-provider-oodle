# Terraform Provider Oodle

This is the repository for the Terraform provider for Oodle.

Manage your monitors and notification policies via terraform!

# Docs
Terraform spec for each resource is documented in [here](docs/resources)

# Example:

Export credentials
```bash
export OODLE_API_KEY="my-api-key"
export OODLE_INSTANCE="my-instance"
export OODLE_DEPLOYMENT=https://us1.oodle.ai/
````

Sample terraform script
```terraform
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
```

# Developers guide

## Lint project
```bash
./goimports.sh
```

## Generate Docs for examples
```bash
make generate
```

## Run tests
```bash
go test ./...
```
