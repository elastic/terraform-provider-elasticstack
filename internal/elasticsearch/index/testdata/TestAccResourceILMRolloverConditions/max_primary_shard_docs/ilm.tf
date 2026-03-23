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
      max_primary_shard_docs = 5000
    }

    readonly {}
  }

  delete {
    delete {}
  }
}
