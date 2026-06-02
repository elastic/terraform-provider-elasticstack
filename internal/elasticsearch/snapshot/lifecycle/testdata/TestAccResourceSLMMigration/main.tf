variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "repo" {
  name = "${var.name}-repo"
  fs {
    location = "/tmp"
  }
}

resource "elasticstack_elasticsearch_snapshot_lifecycle" "test_migration" {
  name       = var.name
  schedule   = "0 30 1 * * ?"
  repository = elasticstack_elasticsearch_snapshot_repository.repo.name
}
