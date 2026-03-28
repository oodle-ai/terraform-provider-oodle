data "oodle_grafana_folders" "all" {}

output "folders" {
  value = data.oodle_grafana_folders.all.folders
}
