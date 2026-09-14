# One oodle_azure_integration resource is one Azure subscription. To
# collect several subscriptions, declare one resource per subscription.
#
# The Azure app registration must already exist and hold the Monitoring
# Reader role on the subscription (or on the resource groups being
# collected). Oodle verifies the credentials against Azure during apply,
# so a missing role assignment fails the apply rather than surfacing
# later as an integration that never starts receiving.

variable "azure_client_secret" {
  description = "Client secret for the Oodle Azure app registration."
  type        = string
  sensitive   = true
}

# Minimal: collect Oodle's default set of services across the whole
# subscription.
resource "oodle_azure_integration" "prod" {
  subscription_name = "Production"
  tenant_id         = "00000000-0000-0000-0000-000000000000"
  subscription_id   = "11111111-1111-1111-1111-111111111111"
  client_id         = "22222222-2222-2222-2222-222222222222"
  client_secret     = var.azure_client_secret
}

# Filtered: pick the services to collect, narrow some of them by resource
# tag, and restrict the whole integration to a few resource groups.
resource "oodle_azure_integration" "staging" {
  subscription_name = "Staging"
  tenant_id         = "00000000-0000-0000-0000-000000000000"
  subscription_id   = "33333333-3333-3333-3333-333333333333"
  client_id         = "22222222-2222-2222-2222-222222222222"
  client_secret     = var.azure_client_secret

  service_filters = [
    {
      service_ids = ["compute_virtualmachines", "compute_virtualmachinescalesets"]
      # Include-only: a resource must carry every tag listed here.
      tags = {
        environment = "staging"
      }
    },
    {
      service_ids = ["storage_storageaccounts", "cache_redis"]
    },
  ]

  resource_groups     = ["rg-staging-east", "rg-staging-west"]
  resource_name_regex = "^stg-"
}
