variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name        = var.name
  description = "All optionals set"
  metadata    = jsonencode({ owner = "test" })

  processors = [
    jsonencode({
      set = {
        field = "_meta"
        value = "indexed"
      }
    }),
  ]

  on_failure = [
    jsonencode({
      set = {
        field = "_index"
        value = "failed-{{ _index }}"
      }
    }),
  ]
}
