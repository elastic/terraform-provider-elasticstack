provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

variable "index_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name          = var.policy_name
  force_destroy = true

  hot {
    min_age = "1h"

    set_priority {
      priority = 10
    }

    rollover {
      max_age = "1d"
    }

    readonly {}
  }
}

resource "elasticstack_elasticsearch_index" "test_index" {
  name                = var.index_name
  deletion_protection = false
  number_of_shards    = 1
  number_of_replicas  = 0
}
