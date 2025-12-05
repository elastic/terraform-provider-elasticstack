variable "connector_name" {
  description = "The connector name"
  type        = string
}

resource "elasticstack_kibana_action_connector" "test" {
  name = "Updated ${var.connector_name}"
  config = jsonencode({
    apiUrl       = "https://bedrock-runtime.us-west-2.amazonaws.com"
    defaultModel = "anthropic.claude-3-5-sonnet-20240620-v1:0"
  })
  secrets = jsonencode({
    accessKey = "updated-access-key"
    secret    = "updated-secret-key"
  })
  connector_type_id = ".bedrock"
}
