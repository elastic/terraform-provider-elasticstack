variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  type        = "query"
  query       = "process.parent.name:(EXCEL.EXE OR WINWORD.EXE OR POWERPNT.EXE OR OUTLOOK.EXE)"
  language    = "kuery"
  enabled     = true
  description = "Detects processes started by MS Office programs"
  severity    = "low"
  risk_score  = 50
  from        = "now-70m"
  to          = "now"
  interval    = "1h"
  index       = ["logs-*", "winlogbeat-*"]

  tags            = ["child process", "ms office", "terraform-test"]
  references      = ["https://attack.mitre.org/techniques/T1566/001/"]
  false_positives = ["Legitimate corporate macros"]
  author          = ["Security Team"]
  license         = "Elastic License v2"
  note            = "Investigate parent process and command line"
  max_signals     = 100

  threat = [
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
}

