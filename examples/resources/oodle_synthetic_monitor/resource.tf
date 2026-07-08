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

# Example: multi-step synthetic monitor.
# Logs in, extracts a token and user id from the response, then calls a
# protected endpoint using those variables.
resource "oodle_synthetic_monitor" "auth_flow" {
  name      = "Auth + Protected API"
  enabled   = true
  rule_type = "multistep"
  interval  = "5m"
  timeout   = "30s"

  rule_config = {
    multistep = {
      steps = [
        {
          name = "Get Token"
          request = {
            url    = "https://api.example.com/auth/token"
            method = "POST"
            headers = {
              "Content-Type" = "application/json"
            }
            body                  = jsonencode({ client_id = "abc", client_secret = "xyz" })
            expected_status_codes = ["2XX"]
          }
          extract = [
            {
              # Extracted values are referenced as {{VAR_NAME}} in later steps.
              name   = "ACCESS_TOKEN"
              source = "body"
              parser = "jsonpath"
              query  = "$.access_token"
              secret = true
            },
            {
              name   = "USER_ID"
              source = "body"
              parser = "jsonpath"
              query  = "$.user.id"
            },
          ]
        },
        {
          name = "Get User Profile"
          request = {
            url          = "https://api.example.com/users/{{USER_ID}}"
            method       = "GET"
            bearer_token = "{{ACCESS_TOKEN}}"

            expected_status_codes = ["200"]
            expected_body         = "\"active\":true"
            max_response_time_ms  = 800
          }
        },
      ]
    }
  }
}
