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
    client                     = "default"
    base_path                  = "snapshots"
    location_mode              = "primary_only"
    compress                   = false
    chunk_size                 = "1gb"
    max_snapshot_bytes_per_sec = "20mb"
    max_restore_bytes_per_sec  = "10mb"
    readonly                   = true
  }
}
