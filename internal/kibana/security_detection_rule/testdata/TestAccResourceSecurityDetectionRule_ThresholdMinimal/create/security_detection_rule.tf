variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "threshold"
  query       = "event.action:login"
  language    = "kuery"
  enabled     = true
  description = "Minimal test threshold security detection rule"
  severity    = "low"
  risk_score  = 21
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]

  threshold = {
    value = 10
    field = ["user.name"]
  }
}

