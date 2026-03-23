variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name = "Updated ${var.connector_name}"
  config = jsonencode({
    apiProvider  = "OpenAI"
    apiUrl       = "https://api.openai.com/v1"
    defaultModel = "gpt-4o"
  })
  secrets = jsonencode({
    apiKey = "updated-api-key"
  })
  connector_type_id = ".gen-ai"
}
