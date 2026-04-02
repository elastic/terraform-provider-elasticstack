provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_hot_actions" {
  name = var.policy_name

  hot {
    min_age = "1h"

    rollover {
      max_age = "7d"
    }

    forcemerge {
      max_num_segments = 1
      index_codec      = "best_compression"
    }

    shrink {
      number_of_shards         = 1
      allow_write_after_shrink = true
    }

    unfollow {
      enabled = true
    }
  }
}
