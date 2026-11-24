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
    every = "1m"
  }

  query = {
    base = "FROM logs-* | STATS count = COUNT(*) BY host.name"
    breach = {
      segment = "WHERE count > 500"
    }
  }

  recovery = {
    strategy = "condition"
    segment  = join("", [for i in range(401) : "aaaaaaaaaaaaaaaaaaaaaaaaa"])
  }

  no_data = {
    strategy = "ignore"
  }
}
