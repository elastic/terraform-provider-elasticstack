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
        value = "dlq-{{ _index }}"
      }
    }),
    jsonencode({
      set = {
        field = "error_reason"
        value = "{{ _ingest.on_failure_message }}"
      }
    }),
  ]
}
