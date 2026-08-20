# A calendar schedule: weekday mornings, held to the wall clock in a
# named timezone across a daylight saving change.
resource "oodle_genai_dataset_schedule" "support_eval_weekdays" {
  dataset_name = oodle_genai_dataset.support_eval.name
  enabled      = true

  mode     = "calendar"
  timezone = "America/Los_Angeles"
  times    = ["09:00"]
  weekdays = ["monday", "tuesday", "wednesday", "thursday", "friday"]

  # What each firing runs. runName is ignored: every firing is
  # numbered on its own.
  experiment_config = jsonencode({
    datasetId       = oodle_genai_dataset.support_eval.dataset_id
    llmConnectionId = oodle_genai_llm_connection.openai.id
    model           = "gpt-4o"
    promptName      = "support-reply"

    # An ordinary judge scores the generation on its own merits.
    evaluatorIds = [oodle_genai_eval_template.answer_resolves_question.id]

    # An output comparer scores it against the dataset item's expected
    # output, and skips an item that has none. The server rejects an
    # id listed under the wrong key.
    outputComparerIds = ["oodle-managed-output-match-v1"]

    evalConnectionId = oodle_genai_llm_connection.openai.id
  })
}

# An interval schedule, for a run whose cadence matters but whose
# wall-clock time does not. At least 5 minutes, at most 365 days.
resource "oodle_genai_dataset_schedule" "regression_every_six_hours" {
  dataset_name = oodle_genai_dataset.regression.name
  enabled      = true

  mode           = "interval"
  interval_value = 6
  interval_unit  = "hours"

  experiment_config = jsonencode({
    datasetId       = oodle_genai_dataset.regression.dataset_id
    llmConnectionId = oodle_genai_llm_connection.openai.id
    promptName      = "support-reply"
  })
}
