variable "name" {
  description = "Unique rule name"
  type        = string
}

variable "rule_id" {
  description = "Rule identifier"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_alerting_v2_rule" "test" {
  rule_id    = var.rule_id
  kind       = "alert"
  time_field = "event.ingested"

  metadata = {
    name         = var.name
    description  = "Hosts whose error rate stays high"
    tags         = ["errors"]
    routing_tags = ["ops"]
  }

  schedule = {
    every    = "1m"
    lookback = "5m"
  }

  query = {
    base = "FROM logs-* | STATS errors = COUNT(*) BY host.name"
    breach = {
      segment = "WHERE errors > 100"
    }
  }

  recovery = {
    strategy = "query"
    query    = "FROM logs-* | STATS errors = COUNT(*) BY host.name | WHERE errors < 10"
  }

  no_data = {
    strategy = "resolve"
    query    = "FROM logs-* | STATS seen = COUNT(*) BY host.name"
  }

  state_transition = {
    pending = {
      count = 2
    }
    recovering = {
      count     = 1
      timeframe = "5m"
      operator  = "or"
    }
  }

  grouping = {
    fields = ["host.name"]
  }

  artifacts = [
    {
      id   = "rb"
      type = "runbook"
      data = jsonencode({ content = "# Check the host" })
    },
  ]
}
