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
  name = var.name

  schedule      = "0 30 2 * * ?"
  snapshot_name = "<daily-snap-{now/d}>"
  repository    = elasticstack_elasticsearch_snapshot_repository.repo.name

  expand_wildcards     = "all"
  indices              = ["data-*", "metrics-*"]
  feature_states       = []
  partial              = false
  ignore_unavailable   = true
  include_global_state = true

  expire_after = "60d"
  min_count    = 3
  max_count    = 30
}
