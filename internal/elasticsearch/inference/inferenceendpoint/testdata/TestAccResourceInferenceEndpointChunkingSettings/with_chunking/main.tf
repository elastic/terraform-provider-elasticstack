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
  task_type    = "text_embedding"
  service      = "openai"
  service_settings = jsonencode({
    api_key    = var.api_key
    model_id   = "text-embedding-3-small"
    dimensions = 128
    similarity = "cosine"
  })
  chunking_settings = jsonencode({
    strategy       = "sentence"
    max_chunk_size = 250
  })
}
