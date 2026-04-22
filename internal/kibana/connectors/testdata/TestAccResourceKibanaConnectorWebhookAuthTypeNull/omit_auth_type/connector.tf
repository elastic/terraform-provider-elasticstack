variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name = var.connector_name
  config = jsonencode({
    url    = "https://hooks.example.com/services"
    method = "post"
    headers = {
      "Content-Type" = "application/json"
    }
    hasAuth = false
  })
  secrets           = jsonencode({})
  connector_type_id = ".webhook"
}
