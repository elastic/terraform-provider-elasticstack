provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_inference" "test" {
  model_id = "my_endpoint"

  input_output {
    input_field  = "foo"
    output_field = "baz"
  }

  field_map = {
    content = "text_field"
    title   = "title_field"
  }

  target_field   = "ml.inference"
  description    = "Run inference on foo"
  if             = "ctx.lang == 'en'"
  ignore_failure = true
  tag            = "inference-tag"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "inference failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "inference_error"
      }
    })
  ]
}
