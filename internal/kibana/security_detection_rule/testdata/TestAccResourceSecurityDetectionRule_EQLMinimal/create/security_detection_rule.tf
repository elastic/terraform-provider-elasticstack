variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "eql"
  query       = "process where process.name == \"cmd.exe\""
  language    = "eql"
  enabled     = true
  description = "Minimal test EQL security detection rule"
  severity    = "low"
  risk_score  = 21
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["winlogbeat-*"]
}

