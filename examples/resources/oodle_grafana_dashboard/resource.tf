# Example of creating a Grafana folder and dashboard together
# The dashboard references the folder's UID

resource "oodle_grafana_folder" "my_folder" {
  title = "Test Folder"
}

resource "oodle_grafana_dashboard" "test_folder" {
  folder = oodle_grafana_folder.my_folder.uid
  config_json = jsonencode({
    "title" : "My Dashboard Title",
    "uid" : "my-dashboard-uid",
    "schemaVersion" : 39,
    "timezone" : "browser",
    "panels" : [
      {
        "id" : 1,
        "type" : "stat",
        "title" : "Request Rate",
        "gridPos" : {
          "h" : 8,
          "w" : 12,
          "x" : 0,
          "y" : 0
        },
        "targets" : [
          {
            "expr" : "sum(rate(http_requests_total[5m]))",
            "refId" : "A"
          }
        ]
      },
      {
        "id" : 2,
        "type" : "timeseries",
        "title" : "Error Rate",
        "gridPos" : {
          "h" : 8,
          "w" : 12,
          "x" : 12,
          "y" : 0
        },
        "targets" : [
          {
            "expr" : "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m]))",
            "refId" : "A"
          }
        ]
      }
    ]
  })
}

# Example of a dashboard with a commit message for version history
resource "oodle_grafana_dashboard" "service_dashboard" {
  folder    = oodle_grafana_folder.my_folder.uid
  message   = "Initial dashboard creation"
  overwrite = true
  config_json = jsonencode({
    "title" : "Service Health",
    "uid" : "service-health-dashboard",
    "schemaVersion" : 39,
    "time" : {
      "from" : "now-6h",
      "to" : "now"
    },
    "panels" : [
      {
        "id" : 1,
        "type" : "gauge",
        "title" : "Uptime",
        "gridPos" : {
          "h" : 8,
          "w" : 8,
          "x" : 0,
          "y" : 0
        },
        "targets" : [
          {
            "expr" : "avg(up{job=\"my-service\"})",
            "refId" : "A"
          }
        ]
      }
    ]
  })
}

# Example without a folder (created at root level)
resource "oodle_grafana_dashboard" "root_dashboard" {
  config_json = jsonencode({
    "title" : "Root Level Dashboard",
    "uid" : "root-level-dashboard",
    "schemaVersion" : 39,
    "panels" : []
  })
}

