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

# ---------------------------------------------------------------------------
# Notifications for synthetic monitors
#
# A synthetic monitor only runs the check and emits the metric
# `oodle_synthetic_monitor_up{monitor_id="<id>"}` (1 = up, 0 = down). It does
# NOT send notifications on its own. To be alerted when a check fails, pair the
# synthetic monitor with a companion `oodle_monitor` that watches this metric
# and routes to your notifiers. This is the same mechanism the Oodle UI uses
# under the hood when you enable notifications on a synthetic monitor.
# ---------------------------------------------------------------------------
resource "oodle_synthetic_monitor" "checkout" {
  name      = "Checkout API"
  enabled   = true
  rule_type = "http"
  interval  = "1m"
  timeout   = "10s"

  rule_config = {
    http = {
      url                   = "https://api.example.com/checkout/health"
      method                = "GET"
      expected_status_codes = ["2XX"]
    }
  }
}

resource "oodle_monitor" "checkout_alert" {
  name = "Synthetic Monitor: Checkout API"

  # Fire when the synthetic monitor's most recent check is failing (up == 0).
  # `monitor_id` is the synthetic monitor's ID, wired up via interpolation.
  promql_query = "oodle_synthetic_monitor_up{monitor_id=\"${oodle_synthetic_monitor.checkout.id}\"} == 0"
  interval     = "1m"

  conditions = {
    critical = {
      operation = "=="
      value     = 0
      # Require the check to stay down for 5m before alerting (tune to taste).
      for = "5m"
    }
  }

  # Link this alert to its synthetic monitor. `_oodle_synthetic_monitor_id`
  # maps the companion alert to the synthetic monitor without reusing its ID,
  # so the Oodle UI's Notifications toggle recognizes this Terraform-managed
  # alert and it is hidden from the regular Monitors list. `source` keeps it
  # consistent with the alerts the Oodle UI creates for synthetic monitors.
  labels = {
    source                      = "synthetic_monitor"
    _oodle_synthetic_monitor_id = oodle_synthetic_monitor.checkout.id
  }

  # Send a single notification for the monitor rather than one per series.
  grouping = {
    by_monitor = true
  }

  annotations = {
    summary = "Synthetic monitor Checkout API is down"
  }

  # Route failures to your notifiers (define these with oodle_notifier).
  notifications = [
    {
      notifiers = {
        any = [oodle_notifier.oncall.id]
      }
    }
  ]
}
