variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_gcs_repo" {
  name   = var.name
  verify = false

  gcs {
    bucket                     = "test-gcs-bucket"
    client                     = "secondary"
    base_path                  = "snapshots/v2"
    compress                   = true
    chunk_size                 = "500mb"
    max_snapshot_bytes_per_sec = "40mb"
    max_restore_bytes_per_sec  = "20mb"
    readonly                   = false
  }
}
