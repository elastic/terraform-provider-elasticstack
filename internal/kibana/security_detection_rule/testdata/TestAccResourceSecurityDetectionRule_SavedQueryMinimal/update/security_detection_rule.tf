variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "saved_query"
  query       = "event.category:authentication"
  enabled     = false
  description = "Updated minimal test saved query security detection rule"
  severity    = "medium"
  risk_score  = 55
  from        = "now-12m"
  to          = "now"
  interval    = "10m"
  index       = ["logs-*", "winlogbeat-*"]
  saved_id    = "test-saved-query-id-updated"
}

