provider "elasticstack" {
  elasticsearch {}
}

// create a repository for snapshots
resource "elasticstack_elasticsearch_snapshot_repository" "repo" {
  name = "my_snap_repo"

  fs {
    location                  = "/tmp/snapshots"
    compress                  = true
    max_restore_bytes_per_sec = "20mb"
  }
}

// create a SLM policy and use the above created repository
resource "elasticstack_elasticsearch_snapshot_lifecycle" "slm_policy" {
  name = "my_slm_policy"

  schedule      = "0 30 1 * * ?"
  snapshot_name = "<daily-snap-{now/d}>"
  repository    = elasticstack_elasticsearch_snapshot_repository.repo.name

  config {
    indices              = ["data-*", "important"]
    ignore_unavailable   = false
    include_global_state = false
  }

  expire_after = "30d"
  min_count    = 5
  max_count    = 50
}
