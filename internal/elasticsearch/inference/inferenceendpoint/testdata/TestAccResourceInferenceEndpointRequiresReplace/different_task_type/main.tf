variable "inference_id" {
  description = "The inference endpoint ID"
  type        = string
}

variable "api_key" {
  description = "The API key used by the inference service"
  type        = string
  sensitive   = true
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_inference_endpoint" "test" {
  inference_id = var.inference_id
  task_type    = "chat_completion"
  service      = "openai"
  service_settings = jsonencode({
    api_key  = var.api_key
    model_id = "gpt-4o-mini"
  })
}
