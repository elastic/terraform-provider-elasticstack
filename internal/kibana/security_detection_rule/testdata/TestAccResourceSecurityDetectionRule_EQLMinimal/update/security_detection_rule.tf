variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "eql"
  query       = "process where process.name == \"powershell.exe\""
  language    = "eql"
  enabled     = true
  description = "Updated minimal test EQL security detection rule"
  severity    = "medium"
  risk_score  = 55
  from        = "now-12m"
  to          = "now"
  interval    = "10m"
  index       = ["winlogbeat-*", "sysmon-*"]
}

