data "oodle_grafana_dashboards" "all" {}

output "dashboards" {
  value = data.oodle_grafana_dashboards.all.dashboards
}
