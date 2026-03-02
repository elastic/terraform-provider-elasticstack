variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "saved_query"
  query       = "*:*"
  enabled     = true
  description = "Minimal test saved query security detection rule"
  severity    = "low"
  risk_score  = 21
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  saved_id    = "test-saved-query-id"
}

