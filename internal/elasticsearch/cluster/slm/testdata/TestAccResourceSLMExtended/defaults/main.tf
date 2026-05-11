variable "name" {
  description = "The SLM policy name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "repo" {
  name = "${var.name}-repo"

  fs {
    location                  = "/tmp/snapshots"
    compress                  = true
    max_restore_bytes_per_sec = "20mb"
  }
}

resource "elasticstack_elasticsearch_snapshot_lifecycle" "test_slm" {
  name       = var.name
  schedule   = "0 30 2 * * ?"
  repository = elasticstack_elasticsearch_snapshot_repository.repo.name
}
