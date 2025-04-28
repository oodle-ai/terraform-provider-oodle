# Example of a resource that creates LogMetrics
resource "oodle_logmetrics" "coverage" {
  name = "tf_app_coverage"

  labels = [
    {
      name  = "environment"
      value = "prod"
    },
    {
      name = "container"
      value_extractor = {
        field = "container_name"
      }
    },
    {
      name = "service",
      value_extractor = {
        field     = "log"
        json_path = "service.id"
      }
    },
    {
      name = "step",
      value_extractor = {
        field = "message"
        regex = "step=(\\w+)"
      }
    }
  ]

  filter = {
    any = [
      {
        all = [
          {
            match = {
              field    = "level"
              operator = "is"
              value    = "error"
            },
          },
          {
            match = {
              field    = "container"
              operator = "matches regex"
              value    = "(checkout|payment)"
            },
          },
          {
            match = {
              field     = "log"
              operator  = "contains"
              json_path = "service.id"
              value     = "123"
            },
          },
          {
            match = {
              field    = "namespace"
              operator = "exists"
            }
          }
        ],
      },
      {
        not = {
          match = {
            field    = "container"
            operator = "is"
            value    = "otel-demo"
          }
        }
      }
    ]
  }

  metric_definitions = [
    {
      name = "oodle_logs_app_log_count"
      type = "count"
    },
    {
      name  = "oodle_logs_app_thread_count"
      type  = "gauge"
      field = "thread_count"
    },
    {
      name      = "oodle_logs_app_duration"
      type      = "histogram"
      field     = "log",
      json_path = "duration"
    },
    {
      name  = "oodle_logs_app_step_count"
      type  = "counter"
      field = "step",
      regex = "step=(\\w+)"
    }
  ]
}
