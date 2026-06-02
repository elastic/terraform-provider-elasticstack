variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name   = var.name
  verify = false

  url {
    url                       = "file:/tmp"
    http_max_retries          = 3
    http_socket_timeout       = "30s"
    compress                  = false
    max_restore_bytes_per_sec = "10mb"
    max_number_of_snapshots   = 100
  }
}
