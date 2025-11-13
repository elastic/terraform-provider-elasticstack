variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

variable "secret_key" {
  description = "The secret key for access"
  type        = string
}

variable "integration_name" {
  description = "The integration name"
  type        = string
  default     = "sql"
}

variable "integration_version" {
  description = "The integration version"
  type        = string
  default     = "1.1.0"
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_policy" {
  name    = var.integration_name
  version = var.integration_version
  force   = true
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "${var.policy_name} Agent Policy"
  namespace       = "default"
  description     = "IntegrationPolicyTest Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  description         = "SQL Integration Policy"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test_policy.name
  integration_version = elasticstack_fleet_integration.test_policy.version

  input {
    input_id = "sql-sql/metrics"
    enabled  = true
    streams_json = jsonencode({
      "sql.sql" : {
        "enabled" : true,
        "vars" : {
          "hosts" : ["root:test@tcp(127.0.0.1:3306)/"],
          "period" : "1m",
          "driver" : "mysql",
          "sql_queries" : "- query: SHOW GLOBAL STATUS LIKE 'Innodb_system%'\n  response_format: variables\n        \n",
          "merge_results" : false,
          "ssl" : "",
          "data_stream.dataset" : "sql",
          "processors" : ""
        }
      }
    })
  }
}