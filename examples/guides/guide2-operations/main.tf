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
  ecom_data_view_id = "ff959d40-b880-11e8-a6d9-e546fe2bba5f"

  ecom_data_source = jsonencode({
    type          = "data_view_spec"
    index_pattern = "kibana_sample_data_ecommerce"
    time_field    = "order_date"
  })
}

resource "elasticstack_kibana_dashboard" "operations" {
  title       = "Operations: eCommerce monitoring"
  description = "Interactive eCommerce operations dashboard with controls, KPIs, trends, and an embedded Discover session."

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

  options = {
    use_margins        = true
    sync_colors        = true
    sync_tooltips      = true
    sync_cursor        = true
    auto_apply_filters = true
    hide_panel_titles  = false
    hide_panel_borders = false
  }

  pinned_panels = [
    {
      type = "options_list_control"
      options_list_control_config = {
        by_field = {
          data_view_id  = local.ecom_data_view_id
          field_name    = "category.keyword"
          title         = "Category"
          single_select = true
          display_settings = {
            placeholder = "Select a category..."
          }
        }
      }
    },
  ]

  panels = [
    {
      type = "vis"
      grid = { x = 0, y = 0, w = 16, h = 10 }
      vis_config = {
        by_value = {
          metric_chart_config = {
            title                 = "Revenue"
            data_source_json      = local.ecom_data_source
            ignore_global_filters = false
            sampling              = 1
            query                 = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                type      = "primary"
                operation = "sum"
                field     = "taxful_total_price"
                format    = { type = "number" }
              })
            }]
          }
        }
      }
    },
    {
      type = "vis"
      grid = { x = 16, y = 0, w = 16, h = 10 }
      vis_config = {
        by_value = {
          metric_chart_config = {
            title                 = "Orders"
            data_source_json      = local.ecom_data_source
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
      grid = { x = 32, y = 0, w = 16, h = 10 }
      vis_config = {
        by_value = {
          metric_chart_config = {
            title                 = "Average order value"
            data_source_json      = local.ecom_data_source
            ignore_global_filters = false
            sampling              = 1
            query                 = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                type      = "primary"
                operation = "average"
                field     = "taxful_total_price"
                format    = { type = "number" }
              })
            }]
          }
        }
      }
    },
    {
      type = "vis"
      grid = { x = 0, y = 10, w = 24, h = 14 }
      vis_config = {
        by_value = {
          xy_chart_config = {
            title = "Orders by category over time"
            axis = {
              y = {
                domain_json = jsonencode({ type = "fit" })
                title       = { value = "Orders", visible = true }
              }
              x = {
                title = { value = "order_date", visible = true }
              }
            }
            decorations = {}
            fitting     = { type = "none" }
            legend      = {}
            query       = { expression = "" }
            layers = [{
              type = "area_stacked"
              data_layer = {
                data_source_json = local.ecom_data_source
                x_json = jsonencode({
                  operation               = "date_histogram"
                  field                   = "order_date"
                  suggested_interval      = "auto"
                  use_original_time_range = false
                  include_empty_rows      = true
                  drop_partial_intervals  = false
                })
                breakdown_by_json = jsonencode({
                  operation = "terms"
                  fields    = ["category.keyword"]
                  limit     = 5
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
      grid = { x = 24, y = 10, w = 24, h = 14 }
      vis_config = {
        by_value = {
          datatable_config = {
            no_esql = {
              title            = "Top categories"
              data_source_json = local.ecom_data_source
              query = {
                language   = "kql"
                expression = ""
              }
              styling = {
                density = { mode = "default" }
                paging  = 10
              }
              rows = [{
                config_json = jsonencode({
                  operation = "terms"
                  fields    = ["category.keyword"]
                  limit     = 10
                  rank_by = {
                    type         = "metric"
                    metric_index = 0
                    direction    = "desc"
                  }
                })
              }]
              metrics = [
                {
                  config_json = jsonencode({
                    operation     = "count"
                    empty_as_null = false
                    format = {
                      type     = "number"
                      compact  = false
                      decimals = 0
                    }
                  })
                },
                {
                  config_json = jsonencode({
                    operation     = "sum"
                    field         = "taxful_total_price"
                    empty_as_null = false
                    format = {
                      type     = "number"
                      compact  = false
                      decimals = 2
                    }
                  })
                },
              ]
              ignore_global_filters = false
              sampling              = 1
            }
          }
        }
      }
    },
    {
      type = "vis"
      grid = { x = 0, y = 24, w = 24, h = 12 }
      vis_config = {
        by_value = {
          pie_chart_config = {
            title                 = "Revenue by category"
            donut_hole            = "m"
            label_position        = "outside"
            data_source_json      = local.ecom_data_source
            ignore_global_filters = false
            sampling              = 1
            query                 = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                operation = "sum"
                field     = "taxful_total_price"
                format    = { type = "number" }
              })
            }]
            group_by = [{
              config_json = jsonencode({
                operation = "terms"
                fields    = ["category.keyword"]
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
    {
      type = "discover_session"
      grid = { x = 24, y = 24, w = 24, h = 12 }
      discover_session_config = {
        title = "Recent orders"
        by_value = {
          tab = {
            dsl = {
              query = {
                expression = ""
                language   = "kql"
              }
              data_source_json = local.ecom_data_source
              column_order = [
                "order_date",
                "products.product_name",
                "taxful_total_price",
                "category",
              ]
              view_mode = "documents"
            }
          }
        }
      }
    },
  ]
}
