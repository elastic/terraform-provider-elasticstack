variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  enabled     = true
  description = "Test query rule with filters removed"
  severity    = "medium"
  risk_score  = 55
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  namespace   = "no-filters-namespace"
}

