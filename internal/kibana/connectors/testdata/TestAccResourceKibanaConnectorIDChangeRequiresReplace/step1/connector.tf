variable "connector_name" {
  description = "The connector name"
  type        = string
}

variable "connector_id" {
  description = "Connector ID"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name         = var.connector_name
  connector_id = var.connector_id
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })
  connector_type_id = ".index"
}
