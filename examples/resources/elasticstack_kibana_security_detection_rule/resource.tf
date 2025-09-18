provider "elasticstack" {
  kibana {}
}

# Basic security detection rule
resource "elasticstack_kibana_security_detection_rule" "example" {
  name        = "Suspicious Activity Detection"
  type        = "query"
  query       = "event.action:logon AND user.name:admin"
  language    = "kuery"
  enabled     = true
  description = "Detects suspicious admin logon activities"
  severity    = "high"
  risk_score  = 75
  from        = "now-6m"
  to          = "now"
  interval    = "5m"

  author          = ["Security Team"]
  tags            = ["security", "authentication", "admin"]
  license         = "Elastic License v2"
  false_positives = ["Legitimate admin access during maintenance windows"]
  references = [
    "https://example.com/security-docs",
    "https://example.com/admin-access-policy"
  ]

  note  = "Investigate the source IP and verify if the admin access is legitimate."
  setup = "Ensure that authentication logs are being collected and indexed."
}

# Advanced security detection rule with custom settings
resource "elasticstack_kibana_security_detection_rule" "advanced" {
  name        = "Advanced Threat Detection"
  type        = "query"
  query       = "process.name:powershell.exe AND process.args:*encoded*"
  language    = "kuery"
  enabled     = true
  description = "Detects encoded PowerShell commands which may indicate malicious activity"
  severity    = "critical"
  risk_score  = 90
  from        = "now-10m"
  to          = "now"
  interval    = "2m"
  max_signals = 200
  version     = 1

  index = [
    "winlogbeat-*",
    "logs-windows-*"
  ]

  author = [
    "Threat Intelligence Team",
    "SOC Analysts"
  ]

  tags = [
    "windows",
    "powershell",
    "encoded",
    "malware",
    "critical"
  ]

  false_positives = [
    "Legitimate encoded PowerShell scripts used by automation",
    "Software installation scripts"
  ]

  references = [
    "https://attack.mitre.org/techniques/T1059/001/",
    "https://example.com/powershell-security-guide"
  ]

  license = "Elastic License v2"
  note    = <<-EOT
    ## Investigation Steps
    1. Examine the full PowerShell command line
    2. Decode any base64 encoded content
    3. Check the parent process that spawned PowerShell
    4. Review network connections made during execution
    5. Check for file system modifications
  EOT

  setup = <<-EOT
    ## Prerequisites
    - Windows endpoint monitoring must be enabled
    - PowerShell logging should be configured
    - Sysmon or equivalent process monitoring required
  EOT
}