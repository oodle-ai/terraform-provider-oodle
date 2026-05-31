# One-shot deploy of an Oodle AWS CloudWatch integration: this module
# provisions the IAM role Oodle assumes (via the same CloudFormation
# template the Oodle UI uses) and registers the integration with Oodle
# in a single terraform apply.
#
# The caller configures the AWS provider this module uses; multi-account
# fleets pass a per-account provider alias via the `providers` argument:
#
#   module "oodle_prod" {
#     source       = "github.com/oodle-ai/terraform-provider-oodle//examples/modules/oodle_aws_integration"
#     external_id  = random_uuid.ext_id.result
#     account_id   = "123456789012"
#     regions      = ["us-west-2", "us-east-1"]
#     namespaces   = ["AWS/EC2", "AWS/RDS", "AWS/Lambda"]
#     providers    = { aws = aws.prod }
#   }
#
# Reuse the same external_id across every module instantiation in a
# workspace so the IAM trust policy is shared.

terraform {
  required_providers {
    aws   = { source = "hashicorp/aws" }
    oodle = { source = "oodle-ai/oodle" }
  }
}

variable "external_id" {
  type        = string
  description = "External ID baked into the IAM role's trust policy. Share across all Oodle AWS integrations in the workspace."
}

variable "account_id" {
  type        = string
  description = "12-digit AWS account ID to monitor."
}

variable "role_name" {
  type        = string
  default     = "OodleIntegrationRole"
  description = "Name of the IAM role created by the CloudFormation stack and assumed by Oodle."
}

variable "oodle_aws_account" {
  type        = string
  default     = "052799302239"
  description = "Oodle's AWS account ID — the principal allowed to assume the IAM role."
}

variable "regions" {
  type        = list(string)
  description = "AWS regions to pull metrics from."
}

variable "namespaces" {
  type        = list(string)
  description = "CloudWatch namespaces to discover (e.g. [\"AWS/EC2\", \"AWS/RDS\"])."
}

variable "cf_region" {
  type        = string
  default     = "us-west-2"
  description = "Region the CloudFormation stack is deployed in."
}

variable "integration_name" {
  type        = string
  default     = null
  description = "Optional human-readable name for the Oodle integration. Server assigns one if omitted."
}

# Permissions granted to Oodle live in the CloudFormation template hosted
# at template_url, not in this Terraform plan. Oodle owns the permission
# set in one place so changes propagate to every customer on the next
# apply; see the template to inspect the trust policy and read-only
# policies it attaches.
resource "aws_cloudformation_stack" "oodle_iam_role" {
  name         = "oodle-aws-integration-role"
  template_url = "https://s3.us-west-2.amazonaws.com/oodle-configs/aws/aws_integration_iam_role.yaml"
  capabilities = ["CAPABILITY_NAMED_IAM"]
  parameters = {
    ExternalId        = var.external_id
    OodleAWSAccountId = var.oodle_aws_account
    RoleName          = var.role_name
  }
}

resource "oodle_aws_integration" "this" {
  account_id             = var.account_id
  role_arn               = "arn:aws:iam::${var.account_id}:role/${var.role_name}"
  external_id            = var.external_id
  regions                = var.regions
  launch_cf_stack_region = var.cf_region
  name                   = var.integration_name

  resource_types_search_tags = [
    {
      resource_types = var.namespaces
    },
  ]

  depends_on = [aws_cloudformation_stack.oodle_iam_role]
}

output "integration_id" {
  value       = oodle_aws_integration.this.id
  description = "Oodle-assigned ID of the AWS integration."
}

output "role_arn" {
  value       = oodle_aws_integration.this.role_arn
  description = "ARN of the IAM role Oodle assumes in the target account."
}
