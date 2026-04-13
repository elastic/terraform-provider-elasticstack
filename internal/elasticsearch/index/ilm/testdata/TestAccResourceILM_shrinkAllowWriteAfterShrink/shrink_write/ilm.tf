provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_shrink" {
  name = var.policy_name

  hot {
    min_age = "1h"
    set_priority {
      priority = 10
    }
    rollover {
      max_age = "1d"
    }
    shrink {
      number_of_shards         = 1
      allow_write_after_shrink = true
    }
  }

  delete {
    delete {}
  }
}
