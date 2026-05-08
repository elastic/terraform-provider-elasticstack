variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_fs_repo" {
  name   = var.name
  verify = false

  fs {
    location                   = "/tmp"
    compress                   = false
    chunk_size                 = "1gb"
    max_snapshot_bytes_per_sec = "20mb"
    max_restore_bytes_per_sec  = "10mb"
    readonly                   = true
    max_number_of_snapshots    = 100
  }
}
