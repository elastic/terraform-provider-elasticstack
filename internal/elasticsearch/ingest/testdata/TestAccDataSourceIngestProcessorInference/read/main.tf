provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_inference" "test" {
  model_id = "my_endpoint"

  input_output {
    input_field  = "foo"
    output_field = "bar"
  }
}
