variable "connector_name" {
  description = "The connector name"
  type        = string
}

variable "auth_token" {
  type      = string
  sensitive = true
  default   = "test-bearer-token-for-issue-1469"
}

resource "elasticstack_kibana_action_connector" "test" {
  name              = var.connector_name
  connector_type_id = ".webhook"
  config = jsonencode({
    hasAuth = false
    headers = {
      "Authorization" = "Bearer ${var.auth_token}"
      "Content-Type"  = "application/json"
    }
    url = "https://example.com/webhook"
  })
  secrets = jsonencode({})
}
