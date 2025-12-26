# Example of creating a Grafana folder
resource "oodle_grafana_folder" "my_folder" {
  title = "My Application Dashboards"
}

# Example of nested folders (folder within a folder)
resource "oodle_grafana_folder" "production" {
  title = "Production"
}

resource "oodle_grafana_folder" "production_services" {
  title      = "Services"
  parent_uid = oodle_grafana_folder.production.uid
}

# Example with a custom UID
resource "oodle_grafana_folder" "platform_team" {
  uid   = "platform-team-dashboards"
  title = "Platform Team"
}

