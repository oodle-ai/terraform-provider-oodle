data "oodle_logmetrics" "all" {}

output "logmetrics" {
  value = data.oodle_logmetrics.all.logmetrics
}
