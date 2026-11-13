provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_alerting_v2_rule" "cpu_high" {
  rule_id    = "cpu-high-prod"
  kind       = "alert"
  time_field = "@timestamp"

  metadata = {
    name        = "CPU usage above 90% (production)"
    description = "Hosts whose average CPU stays above 90% for a minute."
    tags        = ["production", "infra"]
    # Alerts from this rule are picked up by action policies whose matcher.tags contain one of these tags.
    routing_tags = ["pager"]
  }

  schedule = {
    every    = "1m"
    lookback = "5m"
  }

  query = {
    base = "FROM metrics-system.cpu-* | STATS avg_cpu = AVG(system.cpu.total.norm.pct) BY host.name"
    breach = {
      segment = "WHERE avg_cpu > 0.9"
    }
  }

  recovery = {
    strategy = "condition"
    segment  = "WHERE avg_cpu < 0.8"
  }

  no_data = {
    strategy = "keep_last"
    query    = "FROM metrics-system.cpu-* | STATS count = COUNT(*) BY host.name"
  }

  state_transition = {
    pending = {
      count     = 2
      timeframe = "5m"
      operator  = "and"
    }
    recovering = {
      count = 1
    }
  }

  grouping = {
    fields = ["host.name"]
  }

  artifacts = [
    {
      id   = "runbook"
      type = "runbook"
      data = jsonencode({ content = "# CPU high\n\n1. Check the host's top processes." })
    },
    {
      id   = "dashboard"
      type = "dashboard"
      data = jsonencode({ dashboard_id = "hosts-overview" })
    },
  ]
}
