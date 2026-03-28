# Example: Basic HTTP synthetic monitor
resource "oodle_synthetic_monitor" "example" {
  name      = "Example HTTP Monitor"
  enabled   = true
  rule_type = "http"
  interval  = "30s"
  timeout   = "5s"

  rule_config = {
    http = {
      url                   = "https://example.com"
      method                = "GET"
      expected_status_codes = ["2XX"]
      follow_redirects      = true
      insecure_skip_verify  = false
    }
  }
}

# Example: HTTP synthetic monitor with custom headers and body
resource "oodle_synthetic_monitor" "api_check" {
  name      = "API Health Check"
  enabled   = true
  rule_type = "http"
  interval  = "1m"
  timeout   = "10s"

  rule_config = {
    http = {
      url    = "https://api.example.com/health"
      method = "POST"
      headers = {
        "Content-Type"  = "application/json"
        "Authorization" = "Bearer token"
      }
      body                  = "{\"check\": true}"
      expected_status_codes = ["200", "201"]
      follow_redirects      = false
      insecure_skip_verify  = false
    }
  }
}
