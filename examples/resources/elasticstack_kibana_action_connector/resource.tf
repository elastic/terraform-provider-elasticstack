provider "elasticstack" {
  elasticsearch {}
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
    routingKey = "<your-pagerduty-routing-key>"
  })
}

resource "elasticstack_kibana_action_connector" "slack-connector" {
  name              = "slack"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "<your-webhookUrl>"
  })
}

# Slack connector using the Web API method (token based). Requires Kibana 8.8+
# (the .slack_api connector type is not available on earlier versions).
#
# `config.allowedChannels` (available since Kibana 8.11) restricts which Slack
# channels the connector may post to. Each channel always requires a `name`.
# Before Kibana 9.3 the channel `id` is also required; from 9.3 onward `id` is
# optional, so a name-only channel (shown here) is accepted:
#
#   # Kibana 8.11-9.2 (both id and name required):
#   config = jsonencode({
#     allowedChannels = [{ id = "C0123456789", name = "#alerts" }]
#   })
#
# Omit `config` entirely if you do not need to restrict the channels (8.8+).
resource "elasticstack_kibana_action_connector" "slack-api-connector" {
  name              = "slack"
  connector_type_id = ".slack_api"
  config = jsonencode({
    allowedChannels = [
      { name = "#alerts" },
    ]
  })
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
    url      = "<your-webhookUrl>"
    authType = "webhook-authentication-basic"
    hasAuth  = true
    method   = "post"
  })
  secrets = jsonencode({
    user     = "<your-user>"
    password = "<your-password>"
  })
}

resource "elasticstack_kibana_action_connector" "jira-connector" {
  name              = "jira"
  connector_type_id = ".jira"
  config = jsonencode({
    apiUrl     = "https://<your-org>.atlassian.net"
    projectKey = "<your-project-key>"
  })
  secrets = jsonencode({
    email    = "<your-jira-email>"
    apiToken = "<your-jira-api-token>"
  })
}

resource "elasticstack_kibana_action_connector" "teams-connector" {
  name              = "microsoft-teams"
  connector_type_id = ".teams"
  secrets = jsonencode({
    webhookUrl = "<your-teams-webhook-url>"
  })
}
