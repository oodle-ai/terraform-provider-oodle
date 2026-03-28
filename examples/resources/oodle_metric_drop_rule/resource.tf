# Example: Drop all go_gc metrics from an unused exporter
resource "oodle_metric_drop_rule" "drop_go_gc" {
  rule_name = "Drop unused go_gc metrics"
  type      = "series"

  metric_name = {
    name  = "__name__"
    type  = "=~"
    value = "go_gc_.*"
  }

  filters = [
    {
      name  = "job"
      type  = "="
      value = "unused-exporter"
    }
  ]
}

# Example: Drop a specific metric across all jobs
resource "oodle_metric_drop_rule" "drop_specific_metric" {
  rule_name = "Drop kube_state_metrics_total"
  type      = "series"

  metric_name = {
    name  = "__name__"
    type  = "="
    value = "kube_state_metrics_total"
  }
}
