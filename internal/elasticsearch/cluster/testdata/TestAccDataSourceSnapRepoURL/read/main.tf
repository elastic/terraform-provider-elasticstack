variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name = var.name

  url {
    url                        = "file:/tmp"
    http_max_retries           = 3
    http_socket_timeout        = "30s"
    compress                   = true
    max_snapshot_bytes_per_sec = "40mb"
    max_restore_bytes_per_sec  = "10mb"
    readonly                   = false
    max_number_of_snapshots    = 500
    chunk_size                 = "1gb"
  }
}

data "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name = elasticstack_elasticsearch_snapshot_repository.test_url_repo.name
}
