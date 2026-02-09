variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with ES|QL Datatable Panel"
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
      h = 15
    }
    datatable_config = {
      esql = {
        title       = "Sample ESQL Datatable"
        description = "Test ES|QL datatable visualization"
        dataset = jsonencode({
          type  = "esql"
          query = "FROM metrics-* | KEEP host.name, system.cpu.user.pct | LIMIT 10"
        })
        density = {
          mode = "expanded"
        }
        metrics = [
          {
            config = jsonencode({
              operation = "value"
              column    = "system.cpu.user.pct"
              format = {
                id = "number"
                params = {
                  decimals = 2
                }
              }
            })
          }
        ]
        rows = [
          {
            config = jsonencode({
              column      = "host.name"
              operation   = "value"
              collapse_by = "avg"
            })
          }
        ]
        split_metrics_by = [
          {
            config = jsonencode({
              column    = "host.name"
              operation = "value"
            })
          }
        ]
        sort_by = jsonencode({
          column_type = "metric"
          direction   = "desc"
          index       = 0
        })
        ignore_global_filters = false
        sampling              = 1
        paging                = 20
      }
    }
  }]
}
