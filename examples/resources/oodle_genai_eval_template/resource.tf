# An LLM-as-judge template. {{var}} placeholders are filled in by the
# evaluator's variable_mapping.
resource "oodle_genai_eval_template" "answer_resolves_question" {
  name = "answer-resolves-question"
  type = "llm"

  prompt = <<-EOT
    You are grading a support reply.

    Question: {{question}}
    Answer: {{answer}}

    Score 1 if the answer resolves the question, otherwise 0.
  EOT

  vars = ["question", "answer"]

  output_schema = jsonencode({
    score     = "0 or 1"
    reasoning = "one sentence explaining the score"
  })

  model_params = jsonencode({
    temperature = 0
  })
}

# A code scorer. Requires an enterprise plan and the code evaluator
# feature enabled for the instance.
resource "oodle_genai_eval_template" "non_empty_answer" {
  name                 = "non-empty-answer"
  type                 = "code"
  source_code_language = "python"
  vars                 = ["output"]

  source_code = <<-EOT
    def evaluate(output: str, **kwargs) -> dict:
        return {"score": 1.0 if output.strip() else 0.0}
  EOT
}

# An output comparer: a judge with ground truth. Its prompt reads the
# dataset item's expected output, which only exists inside an
# experiment, so it never runs against live traffic and an item with
# no expected output is skipped rather than scored zero.
resource "oodle_genai_eval_template" "matches_expected_answer" {
  name = "matches-expected-answer"
  type = "output_comparer"

  prompt = <<-EOT
    Compare a support reply against the reviewed answer.

    Reply: {{output}}
    Reviewed answer: {{expected_output}}

    Score 1 if the reply says the same thing, otherwise 0.
  EOT

  vars = ["output", "expected_output"]

  output_schema = jsonencode({
    score     = "0 or 1"
    reasoning = "one sentence explaining the score"
  })
}

# A code scorer with settings that imports a shared library. Each
# evaluator on it sets its own values for the settings in params, and
# library_pins fixes the library version it runs. Referencing the
# library's attributes makes Terraform create the library first and
# delete it last.
resource "oodle_genai_eval_template" "house_style_ok" {
  name                 = "house-style-ok"
  type                 = "code"
  source_code_language = "python"

  source_code = <<-EOT
    from oodle_eval.v1 import text
    from shared.house_style import banned_phrases_found

    def evaluate(ctx):
        found = banned_phrases_found(text.reply(ctx), ctx.params["phrases"])
        return EvaluationResult(scores=[
            Score(name="house_style_ok", value=not found, data_type="BOOLEAN"),
        ])
  EOT

  params = jsonencode([
    {
      name        = "phrases"
      type        = "string_list"
      label       = "Banned phrases"
      description = "Phrases the reply must not use."
      default     = ["per my last message"]
    },
  ])

  library_pins = {
    (oodle_genai_code_library.house_style.name) = oodle_genai_code_library.house_style.version
  }
}
