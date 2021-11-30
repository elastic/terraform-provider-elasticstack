terraform {
  required_version = ">= 1.0.0"
  required_providers {
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "~> 0.1.0"
    }
  }
}

provider "elasticstack" {
  elasticsearch {}
}

# create a snapshot repository
resource "elasticstack_elasticsearch_snapshot_repository" "my_fs_repo" {
  name = "my_fs_repo"

  fs {
    location                  = "/tmp"
    compress                  = true
    max_restore_bytes_per_sec = "10mb"
  }
}

# create a snapshot lifecycle policy and use the repository created above to store our snapshots
resource "elasticstack_elasticsearch_snapshot_lifecycle" "my_slm" {
  name = "my_slm_policy"

  schedule      = "0 30 1 * * ?"
  snapshot_name = "<daily-snap-{now/d}>"
  repository    = elasticstack_elasticsearch_snapshot_repository.my_fs_repo.name

  config {
    indices              = ["data-*", "important"]
    ignore_unavailable   = false
    include_global_state = false
  }

  expire_after = "30d"
  min_count    = 5
  max_count    = 50
}
