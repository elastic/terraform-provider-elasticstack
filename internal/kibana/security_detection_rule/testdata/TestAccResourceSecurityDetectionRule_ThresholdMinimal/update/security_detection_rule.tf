variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "threshold"
  query       = "event.action:logout"
  language    = "kuery"
  enabled     = false
  description = "Updated minimal test threshold security detection rule"
  severity    = "medium"
  risk_score  = 55
  from        = "now-12m"
  to          = "now"
  interval    = "10m"
  index       = ["logs-*", "winlogbeat-*"]

  threshold = {
    value = 20
    field = ["host.name"]
  }
}

