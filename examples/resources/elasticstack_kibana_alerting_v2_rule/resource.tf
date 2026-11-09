provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_alerting_v2_rule" "cpu_high" {
  rule_id = "cpu-high-prod"
  kind    = "alert"

  metadata = {
    name = "CPU usage above 90% (production)"
  }

  schedule = {
    every    = "1m"
    lookback = "5m"
  }

  query = {
    base = "FROM metrics-system.cpu-* | STATS avg_cpu = AVG(system.cpu.total.norm.pct) BY host.name | WHERE avg_cpu > 0.9"
  }

  recovery = {
    strategy = "condition"
    segment  = "WHERE avg_cpu < 0.8"
  }

  no_data = {
    strategy = "keep_last"
    query    = "FROM metrics-system.cpu-* | STATS count = COUNT(*) BY host.name"
  }
}
