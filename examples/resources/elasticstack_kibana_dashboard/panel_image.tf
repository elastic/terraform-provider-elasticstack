// Example: image panel with external URL source, presentation fields, and mixed drilldowns.
//
// For uploaded assets, use `image_config.src.file.file_id` with the id returned by Kibana after an upload.
// The provider does not manage file uploads today (`elasticstack_kibana_file` may be added later).

resource "elasticstack_kibana_dashboard" "image_drilldown_target" {
  title            = "Image drilldown target"
  description      = "Target dashboard for the image panel dashboard drilldown"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }
}

resource "elasticstack_kibana_dashboard" "with_image_panel" {
  title            = "Dashboard with image panel"
  description      = "Typed image panel: URL source, object_fit, dashboard + URL drilldowns"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "image"
    grid = { x = 0, y = 0, w = 24, h = 10 }
    image_config = {
      src = {
        url = {
          url = "https://www.elastic.co/favicon.ico"
        }
      }
      alt_text         = "Vendor favicon"
      object_fit       = "contain"
      background_color = "#1d1e31"
      title            = "Branding"
      description      = "Static logo loaded from HTTPS URL"
      hide_title       = false
      hide_border      = true
      drilldowns = [
        {
          dashboard_drilldown = {
            dashboard_id = elasticstack_kibana_dashboard.image_drilldown_target.dashboard_id
            label        = "Open target dashboard"
            trigger      = "on_click_image"
          }
        },
        {
          url_drilldown = {
            url     = "https://example.com/info?q={{context.panel.title}}"
            label   = "External details"
            trigger = "on_click_image"
          }
        },
      ]
    }
  }]

  depends_on = [elasticstack_kibana_dashboard.image_drilldown_target]
}
