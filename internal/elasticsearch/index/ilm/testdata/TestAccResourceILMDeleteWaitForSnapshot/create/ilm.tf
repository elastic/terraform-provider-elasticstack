provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

variable "repository_name" {
  type = string
}

variable "slm_policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_snapshot_repository" "repo" {
  name = var.repository_name

  fs {
    location                  = "/tmp/snapshots"
    compress                  = true
    max_restore_bytes_per_sec = "20mb"
  }
}

resource "elasticstack_elasticsearch_snapshot_lifecycle" "slm" {
  name = var.slm_policy_name

  schedule      = "0 30 1 * * ?"
  snapshot_name = "<daily-snap-{now/d}>"
  repository    = elasticstack_elasticsearch_snapshot_repository.repo.name

  indices              = ["data-*", "abc"]
  ignore_unavailable   = false
  include_global_state = false

  expire_after = "30d"
  min_count    = 5
  max_count    = 50
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_delete_snapshot" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }

  delete {
    wait_for_snapshot {
      policy = elasticstack_elasticsearch_snapshot_lifecycle.slm.name
    }

    delete {
      delete_searchable_snapshot = false
    }
  }
}
