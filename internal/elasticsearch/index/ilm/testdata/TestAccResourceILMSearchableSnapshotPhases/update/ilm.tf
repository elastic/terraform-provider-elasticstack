provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

variable "repository_name" {
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

resource "elasticstack_elasticsearch_index_lifecycle" "test_searchable_snapshot" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }

  cold {
    min_age = "30d"

    searchable_snapshot {
      snapshot_repository = elasticstack_elasticsearch_snapshot_repository.repo.name
    }
  }
}
