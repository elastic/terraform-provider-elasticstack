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
  description = "Test validation with neither index nor data_view_id (should fail)"
  severity    = "medium"
  risk_score  = 50
}

