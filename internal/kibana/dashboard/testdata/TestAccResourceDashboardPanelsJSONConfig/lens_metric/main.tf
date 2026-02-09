variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Panels"
  time_from              = "now-15m"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 0
  query_language         = "kuery"
  query_text             = ""

  panels = [{
    type = "lens"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    config_json = jsonencode({
      "syncColors" : false,
      "syncCursor" : true,
      "syncTooltips" : false,
      "filters" : [],
      "query" : {
        "query" : "",
        "language" : "kuery"
      },
      "attributes" : {
        "title" : "",
        "dataset" : {
          "type" : "dataView",
          "id" : "metrics-*"
        },
        "type" : "metric",
        "sampling" : 1,
        "ignore_global_filters" : false,
        "metrics" : [
          {
            "type" : "primary",
            "operation" : "count",
            "empty_as_null" : true,
            "alignments" : {
              "value" : "right",
              "labels" : "left"
            },
            "fit" : false
          }
        ],
        "query" : {
          "query" : "",
          "language" : "kuery"
        }
      }
    })
  }]
}