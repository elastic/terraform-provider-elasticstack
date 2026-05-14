variable "suffix" {
  type = string
}

variable "dashboard_title" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "slo_idx" {
  name                = "tf-acc-slo-alerts-${var.suffix}-idx"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "slo" {
  name        = "tf-acc-slo-${var.suffix}"
  description = "fixture for slo_alerts dashboard panel acceptance test"
  slo_id      = "tfacc_${var.suffix}"

  kql_custom_indicator {
    index           = elasticstack_elasticsearch_index.slo_idx.name
    filter          = "*"
    good            = "latency < 300"
    total           = "*"
    timestamp_field = "@timestamp"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.95
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  depends_on = [elasticstack_elasticsearch_index.slo_idx]
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with slo_alerts typed panel"

  time_range = {
    from = "now-15m"
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

  panels = [{
    type = "slo_alerts"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    slo_alerts_config = {
      slos = [{
        slo_id = elasticstack_kibana_slo.slo.slo_id
      }]
      title       = "Open violations"
      description = "SLO alerts fixture panel"
      hide_title  = true
      hide_border = false
      drilldowns = [{
        url             = "https://example.com/alerts?slo={{context.panel.title}}"
        label           = "Investigate"
        encode_url      = true
        open_in_new_tab = true
      }]
    }
  }]

  depends_on = [elasticstack_kibana_slo.slo]
}
