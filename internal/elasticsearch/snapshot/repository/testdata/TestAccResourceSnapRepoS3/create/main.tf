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
    bucket                     = "test-bucket"
    endpoint                   = "https://minio.example.com:9000"
    path_style_access          = true
    client                     = "default"
    canned_acl                 = "private"
    storage_class              = "standard"
    server_side_encryption     = false
    base_path                  = "snapshots"
    buffer_size                = "5mb"
    chunk_size                 = "1gb"
    compress                   = false
    max_snapshot_bytes_per_sec = "20mb"
    max_restore_bytes_per_sec  = "10mb"
    readonly                   = true
  }
}
