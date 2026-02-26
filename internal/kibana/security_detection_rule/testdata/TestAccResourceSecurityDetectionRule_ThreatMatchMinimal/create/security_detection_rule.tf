variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "threat_match"
  query       = "destination.ip:*"
  language    = "kuery"
  enabled     = true
  description = "Minimal test threat match security detection rule"
  severity    = "low"
  risk_score  = 21
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]

  threat_index = ["threat-intel-*"]
  threat_query = "threat.indicator.type:ip"

  threat_mapping = [
    {
      entries = [
        {
          field = "destination.ip"
          type  = "mapping"
          value = "threat.indicator.ip"
        }
      ]
    }
  ]
}

