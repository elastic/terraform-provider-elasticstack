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
    bucket                 = "test-bucket"
    endpoint               = "https://minio-alt.example.com:9000"
    path_style_access      = true
    client                 = "default"
    canned_acl             = "private"
    storage_class          = "standard"
    server_side_encryption = false
  }
}
