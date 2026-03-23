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
  description = "Updated detection rule for processes started by MS Office programs"
  severity    = "medium"
  risk_score  = 75
  from        = "now-2h"
  to          = "now"
  interval    = "30m"
  index       = ["logs-*", "winlogbeat-*", "sysmon-*"]

  tags            = ["child process", "ms office", "terraform-test", "updated"]
  references      = ["https://attack.mitre.org/techniques/T1566/001/", "https://attack.mitre.org/techniques/T1204/002/"]
  false_positives = ["Legitimate corporate macros", "Authorized office automation"]
  author          = ["Security Team", "SOC Team"]
  license         = "Elastic License v2"
  note            = "Investigate parent process and command line. Check for malicious documents."
  max_signals     = 200

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
          id        = "T1566"
          name      = "Phishing"
          reference = "https://attack.mitre.org/techniques/T1566"
          subtechnique = [
            {
              id        = "T1566.001"
              name      = "Spearphishing Attachment"
              reference = "https://attack.mitre.org/techniques/T1566/001"
            }
          ]
        },
        {
          id        = "T1204"
          name      = "User Execution"
          reference = "https://attack.mitre.org/techniques/T1204"
          subtechnique = [
            {
              id        = "T1204.002"
              name      = "Malicious File"
              reference = "https://attack.mitre.org/techniques/T1204/002"
            }
          ]
        }
      ]
    }
  ]
}

