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
  description = "Test rule for empty nested list preservation"
  severity    = "low"
  risk_score  = 21
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]

  # Explicit empty lists – should remain [] in state, not become null.
  actions           = []
  exceptions_list   = []
  severity_mapping  = []
  risk_score_mapping = []
  related_integrations = []

  # Non-empty threat with a technique that has an explicit empty subtechnique list.
  threat = [
    {
      framework = "MITRE ATT&CK"
      tactic = {
        id        = "TA0002"
        name      = "Execution"
        reference = "https://attack.mitre.org/tactics/TA0002"
      }
      technique = [
        {
          id           = "T1059"
          name         = "Command and Scripting Interpreter"
          reference    = "https://attack.mitre.org/techniques/T1059"
          subtechnique = []
        }
      ]
    }
  ]
}
