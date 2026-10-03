# Runs an eval template against live traffic. variable_mapping is what
# connects the template's vars to fields on the trace.
resource "oodle_genai_evaluator" "answer_quality" {
  name              = "answer-quality"
  eval_template_id  = oodle_genai_eval_template.answer_resolves_question.id
  llm_connection_id = oodle_genai_llm_connection.openai.id

  enabled                  = true
  target_type              = "trace"
  sampling_rate            = 0.1
  max_invocations_per_hour = 500

  variable_mapping = jsonencode([
    {
      templateVariable = "question"
      langfuseObject   = "trace"
      selectedColumnId = "input"
    },
    {
      templateVariable = "answer"
      langfuseObject   = "trace"
      selectedColumnId = "output"
    },
  ])
}

# An Oodle-managed template can be used instead of your own.
resource "oodle_genai_evaluator" "hallucination" {
  name              = "hallucination"
  eval_template_id  = "oodle-managed-hallucination-v1"
  llm_connection_id = oodle_genai_llm_connection.openai.id
  sampling_rate     = 0.05
}

# An evaluator on a code template. params sets values for the
# template's settings. filters are written in the form the API stores:
# a span:: or resource:: prefix on attribute names, and type 0 (eq),
# 1 (neq), 2 (re), 3 (nre), or "GT", "GTE", "LT", "LTE".
resource "oodle_genai_evaluator" "house_style" {
  name             = "house-style"
  eval_template_id = oodle_genai_eval_template.house_style_ok.id
  enabled          = true
  sampling_rate    = 1

  params = jsonencode({
    phrases = ["per my last message", "as previously stated"]
  })

  filters = jsonencode([
    {
      name  = "span::gen_ai.operation.name"
      type  = 0
      value = "chat"
    },
    {
      name  = "resource::service.name"
      type  = 2
      value = "support-.*"
    },
  ])
}
