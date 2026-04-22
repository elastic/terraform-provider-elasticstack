provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_inference_endpoint" "example" {
  inference_id = "text-embedding-3-large"
  task_type    = "text_embedding"
  service      = "azureaistudio"
  service_settings = jsonencode({
    "api_key" : "example_key",
    "target" : "https://example.com/openai/deployments/text-embedding-3-large/embeddings?api-version=2023-05-151",
    "provider" : "openai",
    "endpoint_type" : "token"
  })
}
