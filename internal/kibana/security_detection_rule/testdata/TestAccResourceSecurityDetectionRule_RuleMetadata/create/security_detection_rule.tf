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
  description = "Test rule for metadata fields coverage"
  severity    = "low"
  risk_score  = 21
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  version     = 1

  setup          = "Install the required integration before enabling this rule."
  timeline_id    = "test-timeline-id-abc123"
  timeline_title = "Test Timeline Template"
}
