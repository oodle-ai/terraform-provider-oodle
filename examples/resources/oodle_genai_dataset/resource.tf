# A dataset is a fixed set of inputs an experiment runs a prompt over.
#
# Datasets have no update endpoint, so changing any attribute replaces
# the dataset — which deletes its items and its run history.
resource "oodle_genai_dataset" "support_eval" {
  name        = "support-eval"
  description = "Questions from the support inbox, with reviewed answers."

  metadata = jsonencode({
    owner = "support-eng"
  })
}
