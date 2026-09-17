# An endpoint you host that runs your own agent. An experiment run
# against it posts every dataset item to the URL and scores the reply.
# Header values are stored encrypted and never returned by the API, so
# keep them out of source control: read them from a variable or a
# secret store.
resource "oodle_genai_webhook" "support_agent" {
  name        = "support-agent-staging"
  description = "LangGraph support triage agent"
  url         = "https://agent-staging.example.com/run"

  headers = {
    Authorization = "Bearer ${var.agent_token}"
  }

  # The body sent per item. A {{path}} placeholder reads the dataset
  # item ({{input}}, {{input.<field>}}, {{metadata.<field>}}, {{id}})
  # and is inserted as JSON. Leave it out to send the item's input as
  # the body.
  request_template = "{\"query\": {{input.question}}, \"channel\": \"eval\"}"

  # Where the output is in the reply. Leave it out to store the whole
  # reply.
  output_path = "result.answer"

  timeout_seconds = 120
}
