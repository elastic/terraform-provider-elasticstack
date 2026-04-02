provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_inference" "inference" {
  model_id = "example"
  input_output {
    input_field  = "body"
    output_field = "body_vector"
  }
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "inference-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_inference.inference.json
  ]
}
