provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_rollover" {
  name = var.policy_name

  hot {
    rollover {
      max_age                = "7d"
      max_docs               = 10000
      max_size               = "100gb"
      max_primary_shard_docs = 5000
      max_primary_shard_size = "50gb"
      min_age                = "3d"
      min_docs               = 1000
      min_size               = "50gb"
      min_primary_shard_docs = 500
      min_primary_shard_size = "25gb"
    }

    readonly {}
  }

  delete {
    delete {}
  }
}
