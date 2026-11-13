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
    name         = "${var.name}-updated"
    description  = "Hosts with an unusual volume of log events"
    tags         = ["z", "a"]
    routing_tags = ["pager", "ops"]
  }

  schedule = {
    every    = "1m"
    lookback = "5m"
  }

  query = {
    base = "FROM logs-* | STATS count = COUNT(*) BY host.name"
    breach = {
      segment = "WHERE count > 500"
    }
  }

  recovery = {
    strategy = "condition"
    segment  = "WHERE count < 50"
  }

  no_data = {
    strategy = "keep_last"
    query    = "FROM logs-* | STATS c = COUNT() BY host.name"
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
      id   = "rb"
      type = "runbook"
      data = jsonencode({ content = "# Steps - check the host first" })
    },
    {
      id   = "db"
      type = "dashboard"
      data = jsonencode({ dashboard_id = "ops-overview" })
    },
  ]
}
