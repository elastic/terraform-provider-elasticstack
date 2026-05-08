variable "name" {
  description = "The snapshot repository name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test_fs_repo" {
  name = var.name

  fs {
    location                   = "/tmp"
    compress                   = true
    readonly                   = false
    max_restore_bytes_per_sec  = "10mb"
    chunk_size                 = "1gb"
    max_snapshot_bytes_per_sec = "20mb"
    max_number_of_snapshots    = 50
  }
}

data "elasticstack_elasticsearch_snapshot_repository" "test_fs_repo" {
  name = elasticstack_elasticsearch_snapshot_repository.test_fs_repo.name
}
