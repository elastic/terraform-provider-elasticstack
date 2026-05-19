variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test" {
  name = var.name

  processors = [
    jsonencode({
      rename = {
        field          = "tmp_source_field"
        target_field   = "destination_field"
        override       = true
        ignore_missing = true
      }
    })
  ]
}
