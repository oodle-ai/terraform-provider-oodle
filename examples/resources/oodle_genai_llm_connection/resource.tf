# The provider credential evaluators and experiments call a model with.
# The key is stored encrypted and is never returned by the API, so keep
# it out of source control — read it from a variable or a secret store.
resource "oodle_genai_llm_connection" "openai" {
  name          = "openai-prod"
  llm_provider  = "openai"
  api_key       = var.openai_api_key
  default_model = "gpt-4o-mini"
  is_default    = true

  custom_models = ["gpt-4o", "gpt-4o-mini"]

  default_params = jsonencode({
    temperature = 0
  })
}
