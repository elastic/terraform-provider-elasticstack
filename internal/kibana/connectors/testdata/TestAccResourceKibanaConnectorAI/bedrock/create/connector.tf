variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name = var.connector_name
  config = jsonencode({
    apiUrl       = "https://bedrock-runtime.us-east-1.amazonaws.com"
    defaultModel = "anthropic.claude-v2"
  })
  secrets = jsonencode({
    accessKey = "test-access-key"
    secret    = "test-secret-key"
  })
  connector_type_id = ".bedrock"
}
