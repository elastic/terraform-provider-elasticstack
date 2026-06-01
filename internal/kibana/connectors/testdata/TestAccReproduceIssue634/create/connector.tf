variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_type_id = ".webhook"
  config = jsonencode({
    url     = "https://hooks.example.com/services"
    method  = "post"
    hasAuth = false
    headers = {
      "Content-Type" = "application/json"
    }
  })
  secrets = jsonencode({})
}
