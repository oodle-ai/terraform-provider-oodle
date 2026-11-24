# Bare oodle_aws_integration resource — use when the IAM role already
# exists in the target AWS account (e.g. deployed via the Oodle UI's
# CloudFormation launch URL, the oodle CLI, or a previous Terraform
# apply). For a one-shot deploy that creates the IAM role and the
# Oodle integration together, see examples/modules/oodle_aws_integration.

# Share the same external_id across every AWS integration in the
# workspace so a single CloudFormation trust policy works for all
# accounts. A random_uuid resource is the simplest way to generate one.
resource "random_uuid" "oodle_external_id" {}

resource "oodle_aws_integration" "prod" {
  account_id  = "123456789012"
  role_arn    = "arn:aws:iam::123456789012:role/OodleIntegrationRole"
  external_id = random_uuid.oodle_external_id.result
  regions     = ["us-west-2", "us-east-1"]

  resource_types_search_tags = [
    {
      resource_types = ["AWS/EC2", "AWS/RDS", "AWS/Lambda"]
      search_tags = [
        { key = "Environment", value = "prod" },
      ]
    },
    {
      resource_types = ["AWS/Logs"]
    },
  ]

  # Optional. Limits the metrics pulled from each namespace, to lower the
  # CloudWatch GetMetricData cost. Patterns match the metric names that
  # Oodle shows, one for each statistic.
  # Omit the attribute to keep the filters set in the Oodle UI; set it to []
  # to remove all filters.
  metric_name_filters = [
    { namespace = "AWS/EC2", include_only = ["aws_ec2_cpuutilization_maximum", "aws_ec2_network*"] },
    { namespace = "AWS/RDS", exclude_only = ["*_samplecount"] },
  ]
}
