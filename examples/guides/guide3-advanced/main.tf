terraform {
  required_providers {
    elasticstack = {
      source = "elastic/elasticstack"
    }
  }
}

provider "elasticstack" {
  kibana {}
}

locals {
  logs_data_source = jsonencode({
    type          = "data_view_spec"
    index_pattern = "kibana_sample_data_logs"
    time_field    = "@timestamp"
  })

  logs_date_histogram_x = jsonencode({
    operation               = "date_histogram"
    field                   = "@timestamp"
    suggested_interval      = "auto"
    use_original_time_range = false
    include_empty_rows      = true
    drop_partial_intervals  = false
  })
}

resource "elasticstack_kibana_dashboard" "advanced" {
  title       = "Advanced: Sections, ES|QL, and access control"
  description = "Production-style dashboard with collapsible sections, ES|QL controls, gauge goals, heatmaps, and write-restricted access."

  time_range = {
    from = "now-7d"
    to   = "now"
  }

  refresh_interval = {
    pause = true
    value = 0
  }

  query = {
    language = "kql"
    text     = ""
  }

  tags = ["advanced", "production"]

  # Uncomment with a non-bootstrap Kibana user that has a profile id (see
  # internal/kibana/dashboard/testdata/TestAccResourceDashboardAccessControl/basic).
  # The default elastic superuser cannot create dashboards with access_mode set.
  # access_control = {
  #   access_mode = "write_restricted"
  # }

  pinned_panels = [
    {
      type = "esql_control"
      esql_control_config = {
        control_type     = "STATIC_VALUES"
        variable_name    = "response_code"
        variable_type    = "values"
        esql_query       = "FROM kibana_sample_data_logs | STATS BY response"
        selected_options = ["200"]
        available_options = [
          "200",
          "404",
          "503",
        ]
        title         = "HTTP response"
        single_select = true
        display_settings = {
          placeholder = "Select response code..."
        }
      }
    },
  ]

  sections = [
    {
      title     = "Activity heatmap"
      collapsed = false
      grid = {
        y = 0
      }
      panels = [
        {
          type = "image"
          grid = { x = 0, y = 0, w = 8, h = 8 }
          image_config = {
            src = {
              url = {
                url = "https://www.elastic.co/favicon.ico"
              }
            }
            alt_text         = "Elastic logo"
            object_fit       = "contain"
            background_color = "#1d1e31"
            title            = "Branding"
            description      = "Static logo for dashboard branding"
            hide_title       = false
            hide_border      = true
          }
        },
        {
          type = "vis"
          grid = { x = 8, y = 0, w = 40, h = 16 }
          vis_config = {
            by_value = {
              heatmap_config = {
                title            = "Requests by hour and response"
                data_source_json = local.logs_data_source
                query = {
                  language   = "kql"
                  expression = ""
                }
                metric_json = jsonencode({
                  operation = "count"
                })
                x_axis_json = jsonencode({
                  operation               = "date_histogram"
                  field                   = "@timestamp"
                  suggested_interval      = "1h"
                  include_empty_rows      = true
                  drop_partial_intervals  = false
                  use_original_time_range = false
                })
                y_axis_json = jsonencode({
                  operation = "terms"
                  fields    = ["response.keyword"]
                  limit     = 8
                  rank_by = {
                    type         = "metric"
                    metric_index = 0
                    direction    = "desc"
                  }
                })
                axis = {
                  x = {
                    labels = {
                      orientation = "horizontal"
                      visible     = true
                    }
                    title = {
                      value   = "Hour of day"
                      visible = true
                    }
                  }
                  y = {
                    labels = {
                      visible = true
                    }
                    title = {
                      value   = "Response code"
                      visible = true
                    }
                  }
                }
                styling = {
                  cells = {
                    labels = {
                      visible = false
                    }
                  }
                }
                legend = {
                  visibility           = "visible"
                  size                 = "m"
                  truncate_after_lines = 5
                }
                ignore_global_filters = false
                sampling              = 1
              }
            }
          }
        },
      ]
    },
    {
      title     = "Goal tracking"
      collapsed = true
      grid = {
        y = 18
      }
      panels = [
        {
          type = "vis"
          grid = { x = 0, y = 0, w = 16, h = 12 }
          vis_config = {
            by_value = {
              gauge_config = {
                title            = "95th percentile bytes"
                data_source_json = local.logs_data_source
                query = {
                  language   = "kql"
                  expression = ""
                }
                metric_json = jsonencode({
                  operation  = "percentile"
                  field      = "bytes"
                  percentile = 95
                  goal = {
                    operation = "static_value"
                    value     = 10000
                  }
                })
                styling = {
                  shape_json = jsonencode({
                    type        = "bullet"
                    orientation = "horizontal"
                  })
                }
                ignore_global_filters = false
                sampling              = 1
              }
            }
          }
        },
        {
          type = "vis"
          grid = { x = 16, y = 0, w = 32, h = 12 }
          vis_config = {
            by_value = {
              xy_chart_config = {
                title = "Bytes sent and request volume"
                axis = {
                  y = {
                    domain_json = jsonencode({ type = "fit" })
                    title       = { value = "Bytes", visible = true }
                  }
                  y2 = {
                    domain_json = jsonencode({ type = "fit" })
                    title       = { value = "Requests", visible = true }
                  }
                  x = {
                    title = { value = "@timestamp", visible = true }
                  }
                }
                decorations = {}
                fitting     = { type = "none" }
                legend = {
                  visibility = "visible"
                  position   = "right"
                  size       = "m"
                  inside     = false
                }
                query = { expression = "" }
                layers = [
                  {
                    type = "area"
                    data_layer = {
                      data_source_json = local.logs_data_source
                      x_json           = local.logs_date_histogram_x
                      y = [{
                        config_json = jsonencode({
                          operation     = "sum"
                          field         = "bytes"
                          empty_as_null = true
                        })
                      }]
                    }
                  },
                  {
                    type = "line"
                    data_layer = {
                      data_source_json = local.logs_data_source
                      x_json           = local.logs_date_histogram_x
                      y = [{
                        config_json = jsonencode({
                          operation     = "count"
                          empty_as_null = true
                          axis          = "y2"
                        })
                      }]
                    }
                  },
                ]
              }
            }
          }
        },
        {
          type = "vis"
          grid = { x = 0, y = 12, w = 48, h = 10 }
          vis_config = {
            by_value = {
              datatable_config = {
                esql = {
                  title = "Requests for selected response"
                  data_source_json = jsonencode({
                    type  = "esql"
                    query = "FROM kibana_sample_data_logs | WHERE response == ?response_code | STATS requests = COUNT(*)"
                  })
                  styling = {
                    density = {
                      mode = "default"
                    }
                  }
                  metrics = [{
                    config_json = jsonencode({
                      operation = "value"
                      column    = "requests"
                    })
                  }]
                  ignore_global_filters = false
                  sampling              = 1
                }
              }
            }
          }
        },
      ]
    },
  ]

  panels = []
}
