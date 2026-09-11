variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_s3_repo" {
  name   = var.name
  verify = false

  s3 {
    bucket                     = "test-bucket-replaced"
    endpoint                   = "https://minio-alt.example.com:9000"
    path_style_access          = false
    client                     = "secondary"
    canned_acl                 = "public-read"
    storage_class              = "reduced_redundancy"
    server_side_encryption     = true
    base_path                  = "snapshots/v2"
    buffer_size                = "10mb"
    chunk_size                 = "500mb"
    compress                   = true
    max_snapshot_bytes_per_sec = "40mb"
    max_restore_bytes_per_sec  = "20mb"
    readonly                   = false
  }
}
