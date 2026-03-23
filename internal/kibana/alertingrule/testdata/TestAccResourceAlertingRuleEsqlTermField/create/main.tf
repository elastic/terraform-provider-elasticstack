variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "esql_term_field" {
  name         = var.name
  rule_type_id = ".es-query"
  consumer     = "alerts"
  enabled      = true
  interval     = "1m"

  params = jsonencode({
    searchType          = "esqlQuery",
    timeWindowSize      = 5,
    timeWindowUnit      = "m",
    threshold           = [0],
    thresholdComparator = ">",
    size                = 100,
    esqlQuery = {
      esql = "FROM logs-* | STATS count = COUNT(*) BY rule.id"
    },
    aggType      = "count",
    groupBy      = "top",
    termField    = "rule.id",
    termSize     = 10,
    timeField    = "@timestamp",
    sourceFields = [],
  })
}
