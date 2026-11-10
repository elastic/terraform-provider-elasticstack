variable "name" {
  description = "Unique rule name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_alerting_v2_rule" "test" {
  kind = "alert"

  metadata = {
    name = var.name
  }

  schedule = {
    every = "1m"
  }

  query = {
    base = "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"
  }

  recovery = {
    strategy = "no_breach"
  }

  no_data = {
    strategy = "ignore"
  }

  artifacts = [
    {
      id   = "rb"
      type = "runbook"
      data = jsonencode(["a"])
    },
  ]
}
