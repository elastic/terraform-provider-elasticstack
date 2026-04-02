variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name        = var.name
  description = "Test Pipeline"
  metadata    = jsonencode({ owner = "test" })

  processors = [
    jsonencode({
      set = {
        description = "My set processor description"
        field       = "_meta"
        value       = "indexed"
      }
    }),
    jsonencode({
      json = {
        field        = "data"
        target_field = "parsed_data"
      }
    }),
  ]

  on_failure = [
    jsonencode({
      set = {
        field = "_index"
        value = "failed-{{ _index }}"
      }
    })
  ]
}
