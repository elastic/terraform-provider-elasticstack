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
    location = "/tmp"
  }
}

data "elasticstack_elasticsearch_snapshot_repository" "test_fs_repo" {
  name = elasticstack_elasticsearch_snapshot_repository.test_fs_repo.name
}
