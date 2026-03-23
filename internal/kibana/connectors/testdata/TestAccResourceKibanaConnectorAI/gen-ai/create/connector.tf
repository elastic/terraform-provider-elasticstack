variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name = var.connector_name
  config = jsonencode({
    apiProvider  = "OpenAI"
    apiUrl       = "https://api.openai.com/v1"
    defaultModel = "gpt-4"
  })
  secrets = jsonencode({
    apiKey = "test-api-key"
  })
  connector_type_id = ".gen-ai"
}
