// Example: `slo_alerts` panel referencing an SLO managed by `elasticstack_kibana_slo`.

resource "elasticstack_elasticsearch_index" "slo_alerts_example_idx" {
  name                = "tf-example-slo-alerts-dashboard"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "slo_alerts_example" {
  name        = "Example SLO for slo_alerts panel"
  description = "Minimal fixture used only as an slo_alerts panel target"
  slo_id      = "tf_example_slo_alerts_panel"

  kql_custom_indicator {
    index           = elasticstack_elasticsearch_index.slo_alerts_example_idx.name
    filter          = "*"
    good            = "latency < 500"
    total           = "*"
    timestamp_field = "@timestamp"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.99
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  depends_on = [elasticstack_elasticsearch_index.slo_alerts_example_idx]
}

resource "elasticstack_kibana_dashboard" "with_slo_alerts_panel" {
  title            = "Dashboard with slo_alerts panel"
  description      = "Typed slo_alerts panel tied to elasticstack_kibana_slo"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "slo_alerts"
    grid = { x = 0, y = 0, w = 24, h = 10 }
    slo_alerts_config = {
      slos = [{
        slo_id = elasticstack_kibana_slo.slo_alerts_example.slo_id
      }]
      title       = "Open violations"
      description = "Alerts tied to the example SLO"
      drilldowns = [{
        url             = "https://example.com/slos/{{context.panel.title}}"
        label           = "Open triage runbook"
        encode_url      = true
        open_in_new_tab = true
      }]
    }
  }]

  depends_on = [elasticstack_kibana_slo.slo_alerts_example]
}
