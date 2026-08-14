# A text prompt. Editing the body publishes a new version rather than
# rewriting the current one, and the "production" label follows the
# version Terraform last applied.
resource "oodle_genai_prompt" "support_reply" {
  name   = "support-reply"
  type   = "text"
  prompt = "Answer the customer question concisely: {{question}}"

  labels         = ["production"]
  tags           = ["support"]
  commit_message = "Ask for a concise answer"

  config = jsonencode({
    model = "gpt-4o-mini"
  })
}

# A chat prompt carries a list of messages instead of a single string.
resource "oodle_genai_prompt" "support_reply_chat" {
  name = "support-reply-chat"
  type = "chat"

  prompt = jsonencode([
    {
      role    = "system"
      content = "You are a terse support agent."
    },
    {
      role    = "user"
      content = "{{question}}"
    },
  ])

  labels = ["production"]
}
