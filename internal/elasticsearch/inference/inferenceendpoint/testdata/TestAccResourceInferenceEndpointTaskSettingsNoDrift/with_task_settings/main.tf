variable "inference_id" {
  description = "The inference endpoint ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_inference_endpoint" "test" {
  inference_id = var.inference_id
  task_type    = "completion"
  service      = "openai"
  service_settings = jsonencode({
    api_key  = "test-openai-api-key"
    model_id = "gpt-4o-mini"
  })
  task_settings = jsonencode({
    user = "test-user"
  })
}
