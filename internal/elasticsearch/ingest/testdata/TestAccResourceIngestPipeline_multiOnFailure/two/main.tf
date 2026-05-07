variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name = var.name

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
    jsonencode({
      set = {
        field = "error_message"
        value = "{{ _ingest.on_failure_message }}"
      }
    }),
  ]
}
