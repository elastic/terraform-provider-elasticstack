// Example: two Discover-session panels — inline (`by_value` DSL tab) and linked (`by_reference`).
//
// For `by_value.tab.dsl`, `tab.esql` is an alternative with ES|QL `data_source_json` instead of DSL fields + filters.
//
// For `by_reference.ref_id`, set this to the Kibana saved object id for your Discover saved search (`search` type on stacks tested here).

resource "elasticstack_kibana_dashboard" "with_discover_session_panels" {
  title            = "Dashboard with discover_session panels"
  description      = "Typed Discover panels: by_value DSL tab + by_reference"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [
    {
      type = "discover_session"
      grid = { x = 0, y = 0, w = 24, h = 12 }
      discover_session_config = {
        title       = "Inline Discover (DSL)"
        description = "by_value.tab.dsl — use ref_id for data_view_reference"
        drilldowns = [{
          url             = "https://example.com/discover?q={{context.panel.title}}"
          label           = "Runbook"
          encode_url      = true
          open_in_new_tab = false
        }]
        by_value = {
          tab = {
            dsl = {
              query = {
                expression = "host.name : *"
              }
              data_source_json = jsonencode({
                type   = "data_view_reference"
                ref_id = "kibana_sample_data_logs"
              })
              column_order = ["@timestamp", "message"]
              view_mode    = "documents"
            }
          }
        }
      }
    },
    {
      type = "discover_session"
      grid = { x = 0, y = 12, w = 24, h = 12 }
      discover_session_config = {
        title = "Saved Discover session"
        by_reference = {
          ref_id = "replace-with-discover-saved-object-id"
        }
      }
    },
  ]
}
