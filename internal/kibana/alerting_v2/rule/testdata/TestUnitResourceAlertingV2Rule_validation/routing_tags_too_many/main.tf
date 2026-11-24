provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_alerting_v2_rule" "test" {
  kind = "alert"

  metadata = {
    name         = "validation"
    routing_tags = [for i in range(21) : "tag_${i}"]
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

  no_data = {
    strategy = "ignore"
  }
}
