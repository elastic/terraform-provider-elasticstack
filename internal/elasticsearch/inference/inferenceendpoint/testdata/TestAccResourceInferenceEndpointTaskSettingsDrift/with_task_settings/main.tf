variable "inference_id" {
  description = "The inference endpoint ID"
  type        = string
}

variable "api_key" {
  description = "The API key used by the inference service"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# Explicitly sets max_new_tokens to a non-default value — this must be surfaced
# as a real change when the previous state had no task_settings.
resource "elasticstack_elasticsearch_inference_endpoint" "test" {
  inference_id = var.inference_id
  task_type    = "completion"
  service      = "azureaistudio"
  service_settings = jsonencode({
    api_key       = var.api_key
    target        = "https://example.com"
    provider      = "openai"
    endpoint_type = "token"
  })
  task_settings = jsonencode({
    max_new_tokens = 32
  })
}
