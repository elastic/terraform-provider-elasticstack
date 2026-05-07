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

  on_failure = []
}
