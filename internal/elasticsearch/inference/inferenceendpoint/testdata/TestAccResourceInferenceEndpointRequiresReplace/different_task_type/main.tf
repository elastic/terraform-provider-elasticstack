variable "inference_id" {
  description = "The inference endpoint ID"
  type        = string
}

variable "url" {
  description = "The Hugging Face inference URL"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_inference_endpoint" "test" {
  inference_id = var.inference_id
  task_type    = "sparse_embedding"
  service      = "hugging_face"
  service_settings = jsonencode({
    api_key = "test-hf-api-key"
    url     = var.url
  })
}
