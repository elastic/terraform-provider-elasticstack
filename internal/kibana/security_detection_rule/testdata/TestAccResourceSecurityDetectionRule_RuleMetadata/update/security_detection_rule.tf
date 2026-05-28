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
  description = "Updated test rule for metadata fields coverage"
  severity    = "medium"
  risk_score  = 50
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  version     = 2

  setup          = "Updated setup instructions with additional prerequisites."
  timeline_id    = "updated-timeline-id-xyz789"
  timeline_title = "Updated Timeline Template"
}
