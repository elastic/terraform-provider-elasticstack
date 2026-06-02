variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_fs_repo" {
  name   = var.name
  verify = true

  fs {
    location                   = "/tmp"
    compress                   = true
    chunk_size                 = "500mb"
    max_snapshot_bytes_per_sec = "40mb"
    max_restore_bytes_per_sec  = "20mb"
    readonly                   = true
    max_number_of_snapshots    = 50
  }
}
