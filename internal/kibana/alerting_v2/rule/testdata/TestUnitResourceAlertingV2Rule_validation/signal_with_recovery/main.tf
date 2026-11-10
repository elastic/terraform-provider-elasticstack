provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_alerting_v2_rule" "test" {
  kind = "signal"

  metadata = {
    name = "validation"
  }

  schedule = {
    every = "1m"
  }

  query = {
    base = "FROM logs-* | STATS count = COUNT(*) BY host.name"
  }

  recovery = {
    strategy = "no_breach"
  }
}
