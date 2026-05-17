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
  description = "Test query rule with all empty list attributes"
  severity    = "low"
  risk_score  = 21
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]

  actions              = []
  exceptions_list      = []
  severity_mapping     = []
  risk_score_mapping   = []
  related_integrations = []
  threat               = [
    {
      framework = "MITRE ATT&CK"
      tactic = {
        id        = "TA0009"
        name      = "Collection"
        reference = "https://attack.mitre.org/tactics/TA0009"
      }
      technique = [
        {
          id        = "T1123"
          name      = "Audio Capture"
          reference = "https://attack.mitre.org/techniques/T1123"
        }
      ]
    }
  ]
  threat_mapping       = []
}
