variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name        = var.name
  description = "Updated Pipeline"
  metadata    = jsonencode({ owner = "updated" })

  processors = [
    jsonencode({
      set = {
        description = "Updated set processor"
        field       = "_meta"
        value       = "reindexed"
      }
    })
  ]
}
