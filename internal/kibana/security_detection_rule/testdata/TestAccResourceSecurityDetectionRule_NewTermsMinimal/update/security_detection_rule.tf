variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                 = var.name
  type                 = "new_terms"
  query                = "host.name:*"
  language             = "kuery"
  enabled              = false
  description          = "Updated minimal test new terms security detection rule"
  severity             = "medium"
  risk_score           = 55
  from                 = "now-12m"
  to                   = "now"
  interval             = "10m"
  index                = ["logs-*", "winlogbeat-*"]
  new_terms_fields     = ["host.name"]
  history_window_start = "now-7d"
}

