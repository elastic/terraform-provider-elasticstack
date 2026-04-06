variable "inference_id" {
  description = "The inference endpoint ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# No task_settings — the API will return {max_new_tokens: 64} as a server default.
# That must not cause drift.
resource "elasticstack_elasticsearch_inference_endpoint" "test" {
  inference_id = var.inference_id
  task_type    = "completion"
  service      = "azureaistudio"
  service_settings = jsonencode({
    api_key       = "test-api-key"
    target        = "https://example.com"
    provider      = "openai"
    endpoint_type = "token"
  })
}
