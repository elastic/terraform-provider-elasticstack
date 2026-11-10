provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_alerting_v2_rule" "test" {
  kind = "alert"

  metadata = {
    name = "validation"
  }

  schedule = {
    every = "1s"
  }

  query = {
    base = "FROM logs-* | STATS count = COUNT(*) BY host.name"
  }

  recovery = {
    strategy = "no_breach"
  }

  no_data = {
    strategy = "ignore"
  }
}
