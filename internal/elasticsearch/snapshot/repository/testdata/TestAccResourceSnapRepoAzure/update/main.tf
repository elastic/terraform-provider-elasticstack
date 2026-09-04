variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_azure_repo" {
  name   = var.name
  verify = false

  azure {
    container                  = "test-azure-container"
    client                     = "secondary"
    base_path                  = "snapshots/v2"
    location_mode              = "secondary_only"
    compress                   = true
    chunk_size                 = "500mb"
    max_snapshot_bytes_per_sec = "40mb"
    max_restore_bytes_per_sec  = "20mb"
    readonly                   = false
  }
}
