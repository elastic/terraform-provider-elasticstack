provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_inference" "test" {
  model_id = "updated_endpoint"

  input_output {
    input_field = "body.content"
  }

  target_field = "ml.updated"
  description  = "Run inference on body.content"
  if           = "ctx.body?.content != null"
  tag          = "updated-inference-tag"
}
