variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "threat_match"
  query       = "source.ip:*"
  language    = "kuery"
  enabled     = false
  description = "Updated minimal test threat match security detection rule"
  severity    = "medium"
  risk_score  = 55
  from        = "now-12m"
  to          = "now"
  interval    = "10m"
  index       = ["logs-*", "winlogbeat-*"]

  threat_index = ["threat-intel-*", "misp-*"]
  threat_query = "threat.indicator.type:domain"

  threat_mapping = [
    {
      entries = [
        {
          field = "source.ip"
          type  = "mapping"
          value = "threat.indicator.domain"
        }
      ]
    }
  ]
}

