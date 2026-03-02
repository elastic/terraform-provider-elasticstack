variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                 = var.name
  type                 = "new_terms"
  query                = "user.name:*"
  language             = "kuery"
  enabled              = true
  description          = "Minimal test new terms security detection rule"
  severity             = "low"
  risk_score           = 21
  from                 = "now-6m"
  to                   = "now"
  interval             = "5m"
  index                = ["logs-*"]
  new_terms_fields     = ["user.name"]
  history_window_start = "now-14d"
}

