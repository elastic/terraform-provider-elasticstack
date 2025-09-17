provider "elasticstack" {
  elasticsearch {}
}

// You can provide the ingest pipeline processors as plain JSON objects.
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

// Or you can use the provided data sources to create the processor data sources.
data "elasticstack_elasticsearch_ingest_processor_set" "set_count" {
  field = "count"
  value = 1
}

data "elasticstack_elasticsearch_ingest_processor_json" "parse_string_source" {
  field        = "string_source"
  target_field = "json_target"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "ingest" {
  name = "set-parse"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_set.set_count.json,
    data.elasticstack_elasticsearch_ingest_processor_json.parse_string_source.json
  ]
}
