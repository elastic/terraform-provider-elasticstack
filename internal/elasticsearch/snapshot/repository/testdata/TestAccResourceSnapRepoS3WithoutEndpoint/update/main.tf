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
    bucket     = "test-bucket"
    chunk_size = "500mb"
  }
}
