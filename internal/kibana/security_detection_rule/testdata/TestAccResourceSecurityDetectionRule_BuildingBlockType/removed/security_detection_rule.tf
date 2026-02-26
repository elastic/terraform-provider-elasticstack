variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = var.name
  type         = "query"
  query        = "process.name:*"
  language     = "kuery"
  enabled      = true
  description  = "Test rule without building block type"
  severity     = "medium"
  risk_score   = 50
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  data_view_id = "no-building-block-data-view-id"
  namespace    = "no-building-block-namespace"
}

