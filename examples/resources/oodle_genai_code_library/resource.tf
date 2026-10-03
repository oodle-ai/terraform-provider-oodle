# A shared Python module. Code templates import it as
# `shared.house_style`. Each change of source_code adds a version.
resource "oodle_genai_code_library" "house_style" {
  name        = "house_style"
  description = "Checks replies against the support style guide."

  source_code = <<-EOT
    from oodle_eval.v1 import text

    def banned_phrases_found(reply, phrases):
        lower = text.normalize(reply)
        return [p for p in phrases if text.normalize(p) in lower]
  EOT
}
