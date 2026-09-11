variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_url_repo" {
  name   = var.name
  verify = true

  url {
    url                        = "file:/tmp/replace"
    http_max_retries           = 7
    http_socket_timeout        = "45s"
    compress                   = true
    max_snapshot_bytes_per_sec = "40mb"
    max_restore_bytes_per_sec  = "20mb"
    readonly                   = false
    max_number_of_snapshots    = 50
  }
}
