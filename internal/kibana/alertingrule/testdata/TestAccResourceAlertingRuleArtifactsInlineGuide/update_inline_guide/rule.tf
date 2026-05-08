variable "name" {
  description = "The rule name"
  type        = string
}

variable "rule_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name     = var.name
  rule_id  = var.rule_id
  consumer = "alerts"
  params = jsonencode({
    aggType             = "avg"
    groupBy             = "top"
    termSize            = 10
    timeWindowSize      = 10
    timeWindowUnit      = "s"
    threshold           = [10]
    thresholdComparator = ">"
    index               = ["test-index"]
    timeField           = "@timestamp"
    aggField            = "version"
    termField           = "name"
  })
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = false

  artifacts {
    investigation_guide {
      content = "# Updated Guide\n\nCheck the metrics."
    }
  }

  lifecycle {
    ignore_changes = [
      last_execution_date,
      last_execution_status,
    ]
  }
}
