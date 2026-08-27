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

# ---------------------------------------------------------------------------
# Network-level checks (ping / tcp / dns / ssl / traceroute)
#
# NOTE: these checks run from Oodle's infrastructure and must target publicly
# reachable hosts. The server rejects internal targets — localhost, private IP
# ranges, and Kubernetes-internal names such as `*.svc`, `*.local`, `*.internal`
# or bare hostnames — because probing them requires routing the check through an
# on-prem agent, which this provider does not yet expose.
# ---------------------------------------------------------------------------

# Example: ICMP ping monitor.
resource "oodle_synthetic_monitor" "ping_check" {
  name      = "Edge Gateway Reachability"
  enabled   = true
  rule_type = "ping"
  interval  = "1m"
  timeout   = "10s"

  rule_config = {
    ping = {
      host = "gateway.example.com"
      # Omit count/interval_ms to use the server defaults (3 packets, 1000ms apart).
      count       = 5
      interval_ms = 500
    }
  }
}

# Example: TCP port connectivity monitor.
resource "oodle_synthetic_monitor" "tcp_check" {
  name      = "Postgres Port Open"
  enabled   = true
  rule_type = "tcp"
  interval  = "1m"
  timeout   = "5s"

  rule_config = {
    tcp = {
      host = "db.example.com"
      port = 5432
    }
  }
}

# Example: DNS resolution monitor.
resource "oodle_synthetic_monitor" "dns_check" {
  name      = "MX Records Present"
  enabled   = true
  rule_type = "dns"
  interval  = "5m"
  timeout   = "10s"

  rule_config = {
    dns = {
      domain      = "example.com"
      record_type = "MX"
      # The check fails unless the lookup returns one of these records.
      expected_values = ["mail.example.com (priority: 10)"]
      # Query a specific resolver instead of the system one.
      nameserver        = "8.8.8.8:53"
      expect_resolution = true
    }
  }
}

# Example: SSL certificate expiry monitor.
resource "oodle_synthetic_monitor" "ssl_check" {
  name      = "API Certificate Expiry"
  enabled   = true
  rule_type = "ssl"
  interval  = "1h"
  timeout   = "10s"

  rule_config = {
    ssl = {
      host = "api.example.com"
      port = 443
      # Fail the check as the certificate approaches expiry.
      warn_days_before_expiry     = 30
      critical_days_before_expiry = 7
      check_certificate_authority = true
    }
  }
}

# Example: traceroute monitor for network path visibility.
resource "oodle_synthetic_monitor" "traceroute_check" {
  name      = "Path To Upstream API"
  enabled   = true
  rule_type = "traceroute"
  interval  = "5m"
  timeout   = "30s"

  rule_config = {
    traceroute = {
      host = "api.example.com"
      # Omit to use the server defaults (30 hops, 1000ms per hop).
      max_hops           = 20
      timeout_per_hop_ms = 1500
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
