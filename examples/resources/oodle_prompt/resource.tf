# A prompt an application fetches by name. Changing the text
# publishes a new version and moves the label to it.
resource "oodle_prompt" "support_reply" {
  name           = "support/reply"
  commit_message = "Ask for the order number first"
  prompt         = <<-EOT
    You are a support assistant for an online store.
    Ask for the order number before anything else.
  EOT
}

# One prompt can reference another. The reference is stored,
# not a copy, so the shared text is edited in one place.
resource "oodle_prompt" "support_base" {
  name   = "support/base"
  prompt = "Be brief. Never promise a refund."
}

resource "oodle_prompt" "support_escalation" {
  name = "support/escalation"
  prompt = join("\n", [
    "@@@oodlePrompt:name=support/base|label=production@@@",
    "Escalate to a person if the customer asks twice.",
  ])
  depends_on = [oodle_prompt.support_base]
}
