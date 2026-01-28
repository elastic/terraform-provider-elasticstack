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
      "enhancements" : {
        "dynamicActions" : {
          "events" : []
        }
      },
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
        "visualizationType" : "lnsMetric",
        "type" : "lens",
        "references" : [
          {
            "type" : "index-pattern",
            "id" : "logs-*",
            "name" : "indexpattern-datasource-layer-b6744a6e-8ae5-4242-a867-67fe650a49fd"
          }
        ],
        "state" : {
          "visualization" : {
            "layerId" : "b6744a6e-8ae5-4242-a867-67fe650a49fd",
            "layerType" : "data",
            "metricAccessor" : "d436609a-d400-473e-aa00-63235b265de1",
            "secondaryTrend" : {
              "type" : "none"
            },
            "secondaryLabelPosition" : "before"
          },
          "query" : {
            "query" : "",
            "language" : "kuery"
          },
          "filters" : [],
          "datasourceStates" : {
            "formBased" : {
              "layers" : {
                "b6744a6e-8ae5-4242-a867-67fe650a49fd" : {
                  "columns" : {
                    "d436609a-d400-473e-aa00-63235b265de1" : {
                      "label" : "Count of records",
                      "dataType" : "number",
                      "operationType" : "count",
                      "isBucketed" : false,
                      "sourceField" : "___records___",
                      "params" : {
                        "emptyAsNull" : true
                      }
                    }
                  },
                  "columnOrder" : [
                    "d436609a-d400-473e-aa00-63235b265de1"
                  ],
                  "incompleteColumns" : {},
                  "sampling" : 1
                }
              }
            },
            "indexpattern" : {
              "layers" : {}
            },
            "textBased" : {
              "layers" : {}
            }
          },
          "internalReferences" : [],
          "adHocDataViews" : {}
        },
        "version" : 1
      }
    })
  }]
}
