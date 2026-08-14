# One row of a dataset.
resource "oodle_genai_dataset_item" "password_reset" {
  dataset_name = oodle_genai_dataset.support_eval.name

  input           = jsonencode({ question = "How do I reset my password?" })
  expected_output = jsonencode({ answer = "Use the reset link on the login page." })

  metadata = jsonencode({
    reviewed_by = "support-eng"
  })
}

# Items are ordinary resources, so a whole dataset can be built from a
# list without repeating the block.
locals {
  cases = [
    {
      question = "Where are my invoices?"
      answer   = "Billing > Invoices in the sidebar."
    },
    {
      question = "How do I add a teammate?"
      answer   = "Settings > Members > Invite."
    },
  ]
}

resource "oodle_genai_dataset_item" "cases" {
  for_each = { for case in local.cases : case.question => case }

  dataset_name    = oodle_genai_dataset.support_eval.name
  input           = jsonencode({ question = each.value.question })
  expected_output = jsonencode({ answer = each.value.answer })
}
