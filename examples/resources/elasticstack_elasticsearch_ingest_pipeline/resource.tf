provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name        = "my_ingest_pipeline"
  description = "My first ingest pipeline managed by Terraform"

  // processors can be defined in different way
  processors = [
    // using the jsonencode function, which is the recommended way if you want to provide JSON object by yourself
    jsonencode({
      set = {
        description = "My set processor description"
        field       = "_meta"
        value       = "indexed"
      }
    }),
    // or use the HERE DOC construct to provide the processor definition
    <<EOF
    {"json": {
      "field": "data",
      "target_field": "parsed_data"
    }}
EOF
    ,
  ]
}
