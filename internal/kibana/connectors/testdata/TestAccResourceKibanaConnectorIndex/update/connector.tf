variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name = "Updated ${var.connector_name}"
  config = jsonencode({
    index   = ".kibana"
    refresh = false
  })
  connector_type_id = ".index"
}
