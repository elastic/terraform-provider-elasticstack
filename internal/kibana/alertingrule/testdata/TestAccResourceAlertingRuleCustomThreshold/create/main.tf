variable "name" {
  type = string
}

resource "elasticstack_kibana_alerting_rule" "custom_threshold" {
  name         = var.name
  rule_type_id = "observability.rules.custom_threshold"
  consumer     = "logs"
  enabled      = false
  interval     = "1m"
  notify_when  = "onActiveAlert"

  params = jsonencode({
    criteria = [
      {
        comparator = ">",
        metrics = [
          {
            name    = "A",
            aggType = "count"
          }
        ],
        threshold = [1],
        timeSize  = 5,
        timeUnit  = "m"
      }
    ],
    searchConfiguration = {
      index = "logs-*",
      query = {
        language = "kuery",
        query    = ""
      }
    }
  })
}
