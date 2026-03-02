variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = var.name
  type         = "query"
  query        = "*:*"
  language     = "kuery"
  enabled      = true
  description  = "Test validation with both index and data_view_id (should fail)"
  severity     = "medium"
  risk_score   = 50
  index        = ["logs-*"]
  data_view_id = "test-data-view-id"
}

