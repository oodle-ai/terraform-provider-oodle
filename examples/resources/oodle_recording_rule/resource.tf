# Example: Record a request rate that dashboards read often
resource "oodle_recording_rule" "checkout_request_rate" {
  name            = "Checkout request rate"
  metric_name     = "service:http_requests:rate5m"
  prometheus_expr = "sum by (service) (rate(http_requests_total{namespace=\"checkout\"}[5m]))"

  labels = {
    team = "checkout"
  }
}

# Example: SLO error ratios that run in order in one execution group
resource "oodle_recording_rule" "postbox_error_ratio_5m" {
  name            = "postbox api availability 5m"
  metric_name     = "slo:sli_error:ratio_rate5m"
  execution_group = "postbox-api-availability"
  prometheus_expr = <<-EOT
    sum(rate(http_requests_total{service="postbox",code=~"5.."}[5m]))
    /
    sum(rate(http_requests_total{service="postbox"}[5m]))
  EOT

  labels = {
    sloth_id = "postbox-api-availability"
  }
}
