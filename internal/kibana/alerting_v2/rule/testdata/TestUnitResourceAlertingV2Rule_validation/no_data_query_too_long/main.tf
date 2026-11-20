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
    strategy = "no_breach"
  }

  no_data = {
    strategy = "keep_last"
    query    = join("", [for i in range(401) : "aaaaaaaaaaaaaaaaaaaaaaaaa"])
  }
}
