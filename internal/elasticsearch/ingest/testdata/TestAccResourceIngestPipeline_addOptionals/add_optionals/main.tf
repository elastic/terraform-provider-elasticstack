variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name        = var.name
  description = "Added via update"
  metadata    = jsonencode({ added = true })

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
        field = "error"
        value = "{{ _ingest.on_failure_message }}"
      }
    }),
  ]
}
