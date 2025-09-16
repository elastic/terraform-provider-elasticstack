---
subcategory: "Kibana"
layout: ""
page_title: "Elasticstack: elasticstack_kibana_security_detection_rule Resource"
description: |-
  Creates or updates a Kibana security detection rule.
---

# Resource: elasticstack_kibana_security_detection_rule

Creates or updates a Kibana security detection rule. Security detection rules are used to detect suspicious activities and generate security alerts based on specified conditions and queries.

See the [Elastic Security detection rules documentation](https://www.elastic.co/guide/en/security/current/rules-api-create.html) for more details.

Note that this Terraform resource only supports Kibana versions >= 8.11.0 

## Example Usage

### Basic Detection Rule

```terraform
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
```

## Argument Reference

The following arguments are supported:

### Required Arguments

- `name` - (String) A human-readable name for the rule.
- `query` - (String) The query language definition used to detect events.
- `description` - (String) The rule's description explaining what it detects.

### Optional Arguments

- `space_id` - (String) An identifier for the space. If not provided, the default space is used. **Note**: Changing this forces a new resource to be created.
- `rule_id` - (String) A stable unique identifier for the rule object. If omitted, a UUID is generated. **Note**: Changing this forces a new resource to be created.
- `type` - (String) Rule type. Currently only `query` is supported. Defaults to `"query"`.
- `language` - (String) The query language (`kuery` or `lucene`). Defaults to `"kuery"`.
- `enabled` - (Boolean) Determines whether the rule is enabled. Defaults to `true`.
- `severity` - (String) Severity level of alerts (`low`, `medium`, `high`, `critical`). Defaults to `"medium"`.
- `risk_score` - (Number) A numerical representation of the alert's severity from 0 to 100. Defaults to `50`.
- `from` - (String) Time from which data is analyzed using date math range (e.g., `now-6m`). Defaults to `"now-6m"`.
- `to` - (String) Time to which data is analyzed using date math range. Defaults to `"now"`.
- `interval` - (String) Frequency of rule execution using date math range (e.g., `5m`). Defaults to `"5m"`.
- `index` - (List of String) Indices on which the rule functions. Defaults to Security Solution default indices.
- `author` - (List of String) The rule's author(s).
- `tags` - (List of String) Tags to help categorize, filter, and search rules.
- `license` - (String) The rule's license.
- `false_positives` - (List of String) Common reasons why the rule may issue false-positive alerts.
- `references` - (List of String) References and URLs to sources of additional information.
- `note` - (String) Notes to help investigate alerts produced by the rule.
- `setup` - (String) Setup guide with instructions on rule prerequisites.
- `max_signals` - (Number) Maximum number of alerts the rule can create during a single run. Defaults to `100`.
- `version` - (Number) The rule's version number. Defaults to `1`.

### Read-Only Attributes

- `id` - (String) The internal identifier of the resource in the format `space_id/rule_object_id`.
- `created_at` - (String) The time the rule was created.
- `created_by` - (String) The user who created the rule.
- `updated_at` - (String) The time the rule was last updated.
- `updated_by` - (String) The user who last updated the rule.
- `revision` - (Number) The rule's revision number representing the version of the rule object.

## Import

Security detection rules can be imported using the rule's object ID:

```shell
terraform import elasticstack_kibana_security_detection_rule.example default/12345678-1234-1234-1234-123456789abc
```

**Note**: When importing, you may need to adjust the `space_id` in your configuration to match the space where the rule was created.