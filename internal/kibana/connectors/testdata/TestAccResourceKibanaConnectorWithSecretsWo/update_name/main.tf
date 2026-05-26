variable "connector_name" {
  description = "The connector name"
  type        = string
}

variable "routing_key" {
  description = "The PagerDuty routing key"
  ephemeral   = true
  type        = string
}

variable "secrets_version" {
  description = "The version identifier for secrets_wo"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_type_id = ".pagerduty"
  config = jsonencode({
    apiUrl = "https://events.pagerduty.com/v2/enqueue"
  })
  secrets_wo         = jsonencode({ routingKey = var.routing_key })
  secrets_wo_version = var.secrets_version
}
