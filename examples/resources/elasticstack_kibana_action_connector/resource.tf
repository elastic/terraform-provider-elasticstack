provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "example" {
  name = "%s"
  config = jsonencode({
    index   = ".kibana"
    refresh = true
  })
  connector_type_id = ".index"
}

resource "elasticstack_kibana_action_connector" "pagerduty-connector" {
  name              = "pagerduty"
  connector_type_id = ".pagerduty"
  config = jsonencode({
    apiUrl = "https://events.pagerduty.com/v2/enqueue"
  })
  secrets = jsonencode({
    routingKey = pagerduty_service_integration.kibana.integration_key
  })
}

resource "elasticstack_kibana_action_connector" "slack-connector" {
  name              = "slack"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "<your-webhookUrl>"
  })
}

resource "elasticstack_kibana_action_connector" "slack-api-connector" {
  name              = "slack"
  connector_type_id = ".slack_api"
  secrets = jsonencode({
    token = "<your-token>"
  })
}

resource "elasticstack_kibana_action_connector" "bedrock-connector" {
  name              = "aws-bedrock"
  connector_type_id = ".bedrock"
  config = jsonencode({
    apiUrl       = "https://bedrock-runtime.us-east-1.amazonaws.com"
    defaultModel = "anthropic.claude-v2"
  })
  secrets = jsonencode({
    accessKey = "<your-aws-access-key>"
    secret    = "<your-aws-secret-key>"
  })
}

resource "elasticstack_kibana_action_connector" "genai-openai-connector" {
  name              = "openai"
  connector_type_id = ".gen-ai"
  config = jsonencode({
    apiProvider  = "OpenAI"
    apiUrl       = "https://api.openai.com/v1"
    defaultModel = "gpt-4"
  })
  secrets = jsonencode({
    apiKey = "<your-openai-api-key>"
  })
}

resource "elasticstack_kibana_action_connector" "genai-azure-connector" {
  name              = "azure-openai"
  connector_type_id = ".gen-ai"
  config = jsonencode({
    apiProvider = "Azure OpenAI"
    apiUrl      = "https://my-resource.openai.azure.com/openai/deployments/my-deployment"
  })
  secrets = jsonencode({
    apiKey = "<your-azure-api-key>"
  })
}

resource "elasticstack_kibana_action_connector" "webhook" {
  name              = "webhook"
  connector_type_id = ".webhook"
  config = jsonencode({
    url      = "<your-webhookUrl>",
    authType = "webhook-authentication-basic",
    hasAuth  = true,
    method   = "post"
  })
  secrets = jsonencode({
    user     = "<your-user>"
    password = "<your-password>"
  })
}
