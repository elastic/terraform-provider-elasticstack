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
  time_field = "event.created"

  metadata = {
    name         = var.name
    description  = "Services whose error rate stays high"
    tags         = ["errors", "services"]
    routing_tags = ["pager"]
  }

  schedule = {
    every    = "1m"
    lookback = "10m"
  }

  query = {
    base = "FROM logs-* | STATS errors = COUNT(*) BY host.name, service.name"
    breach = {
      segment = "WHERE errors > 250"
    }
  }

  recovery = {
    strategy = "manual"
  }

  no_data = {
    strategy = "resolve"
  }

  state_transition = {
    pending = {
      count     = 3
      timeframe = "10m"
      operator  = "or"
    }
  }

  grouping = {
    fields = ["host.name", "service.name"]
  }

  artifacts = [
    {
      id   = "db"
      type = "dashboard"
      data = jsonencode({ dashboard_id = "service-errors" })
    },
  ]
}
