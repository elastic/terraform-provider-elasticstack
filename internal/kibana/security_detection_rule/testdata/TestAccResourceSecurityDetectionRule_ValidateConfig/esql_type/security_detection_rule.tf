variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "esql"
  query       = "FROM logs-* | WHERE event.action == \"login\" | STATS count(*) BY user.name"
  language    = "esql"
  enabled     = true
  description = "Test ESQL validation bypass - neither index nor data_view_id required"
  severity    = "medium"
  risk_score  = 50
}

