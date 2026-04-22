provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_warm_actions" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }

  warm {
    min_age = "30d"

    allocate {
      number_of_replicas    = 2
      total_shards_per_node = 5
      include = jsonencode({
        box_type = "warm"
      })
      require = jsonencode({
        storage = "fast"
      })
    }

    downsample {
      fixed_interval = "1d"
      wait_timeout   = "12h"
    }

    shrink {
      number_of_shards         = 1
      allow_write_after_shrink = true
    }
  }
}
