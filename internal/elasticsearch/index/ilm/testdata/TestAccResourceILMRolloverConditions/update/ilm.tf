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
      max_age                = "14d"
      max_docs               = 15000
      max_size               = "150gb"
      max_primary_shard_docs = 8000
      max_primary_shard_size = "75gb"
      min_age                = "5d"
      min_docs               = 2000
      min_size               = "60gb"
      min_primary_shard_docs = 750
      min_primary_shard_size = "30gb"
    }

    readonly {}
  }

  delete {
    delete {}
  }
}
