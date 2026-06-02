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

resource "elasticstack_elasticsearch_snapshot_lifecycle" "test_slm_metadata" {
  name = var.name

  schedule      = "0 30 1 * * ?"
  snapshot_name = "<daily-snap-{now/d}>"
  repository    = elasticstack_elasticsearch_snapshot_repository.repo.name

  indices              = ["data-*", "abc"]
  ignore_unavailable   = false
  include_global_state = false

  expire_after = "30d"
  min_count    = 5
  max_count    = 50

  metadata = jsonencode({
    created_by = "terraform"
    purpose    = "daily backup"
  })
}
