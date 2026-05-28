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
}

resource "elasticstack_kibana_dashboard" "getting_started" {
  title       = "Getting started: Web server logs"
  description = "A step-by-step web server log monitoring dashboard built with Terraform using Kibana sample logs data."

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

  panels = [
    {
      type = "markdown"
      grid = { x = 0, y = 0, w = 48, h = 6 }
      markdown_config = {
        by_value = {
          title   = "About this dashboard"
          content = <<-EOT
            # Welcome

            This dashboard monitors **web server logs** from the Kibana sample logs dataset.
            Explore request volume, top URLs, and HTTP response codes over the last 7 days.

            Learn more in the [Kibana dashboards guide](https://www.elastic.co/docs/explore-analyze/dashboards).
          EOT
          settings = {
            open_links_in_new_tab = true
          }
        }
      }
    },
    {
      type = "vis"
      grid = { x = 0, y = 6, w = 24, h = 10 }
      vis_config = {
        by_value = {
          metric_chart_config = {
            title                 = "Total events"
            data_source_json      = local.logs_data_source
            ignore_global_filters = false
            sampling              = 1
            query                 = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                type      = "primary"
                operation = "count"
                format    = { type = "number" }
              })
            }]
          }
        }
      }
    },
    {
      type = "vis"
      grid = { x = 24, y = 6, w = 24, h = 10 }
      vis_config = {
        by_value = {
          metric_chart_config = {
            title                 = "Unique client IPs"
            data_source_json      = local.logs_data_source
            ignore_global_filters = false
            sampling              = 1
            query                 = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                type      = "primary"
                operation = "unique_count"
                field     = "clientip"
                format    = { type = "number" }
              })
            }]
          }
        }
      }
    },
    {
      type = "vis"
      grid = { x = 0, y = 16, w = 48, h = 14 }
      vis_config = {
        by_value = {
          xy_chart_config = {
            title = "Request volume over time"
            axis = {
              y = {
                domain_json = jsonencode({ type = "fit" })
                title       = { value = "Count", visible = true }
              }
              x = {
                title = { value = "@timestamp", visible = true }
              }
            }
            decorations = {}
            fitting     = { type = "none" }
            legend      = {}
            query       = { expression = "" }
            layers = [{
              type = "line"
              data_layer = {
                data_source_json = local.logs_data_source
                x_json = jsonencode({
                  operation               = "date_histogram"
                  field                   = "@timestamp"
                  suggested_interval      = "auto"
                  use_original_time_range = false
                  include_empty_rows      = true
                  drop_partial_intervals  = false
                })
                y = [{
                  config_json = jsonencode({
                    operation     = "count"
                    empty_as_null = true
                  })
                }]
              }
            }]
          }
        }
      }
    },
    {
      type = "vis"
      grid = { x = 0, y = 30, w = 24, h = 14 }
      vis_config = {
        by_value = {
          xy_chart_config = {
            title = "Top 10 URLs"
            axis = {
              y = { domain_json = jsonencode({ type = "fit" }) }
            }
            decorations = {
              minimum_bar_height = 1
              show_value_labels  = false
            }
            fitting = { type = "none" }
            legend  = {}
            query   = { expression = "" }
            layers = [{
              type = "bar_horizontal"
              data_layer = {
                data_source_json = local.logs_data_source
                x_json = jsonencode({
                  operation = "terms"
                  fields    = ["url.keyword"]
                  limit     = 10
                  rank_by = {
                    type         = "metric"
                    metric_index = 0
                    direction    = "desc"
                  }
                })
                y = [{
                  config_json = jsonencode({
                    operation     = "count"
                    empty_as_null = true
                  })
                }]
              }
            }]
          }
        }
      }
    },
    {
      type = "vis"
      grid = { x = 24, y = 30, w = 24, h = 14 }
      vis_config = {
        by_value = {
          pie_chart_config = {
            title                 = "Response codes"
            donut_hole            = "s"
            label_position        = "outside"
            data_source_json      = local.logs_data_source
            ignore_global_filters = false
            sampling              = 1
            query                 = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                operation = "count"
                format    = { type = "number" }
              })
            }]
            group_by = [{
              config_json = jsonencode({
                operation = "terms"
                fields    = ["response.keyword"]
                limit     = 10
                rank_by = {
                  type         = "metric"
                  metric_index = 0
                  direction    = "desc"
                }
                color = {
                  mode    = "categorical"
                  palette = "default"
                  mapping = []
                }
              })
            }]
          }
        }
      }
    },
  ]
}
