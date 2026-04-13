provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_cold_actions" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }

  cold {
    min_age = "30d"

    allocate {
      number_of_replicas    = 2
      total_shards_per_node = 4
      include = jsonencode({
        box_type = "cold"
      })
      exclude = jsonencode({
        box_type = "warm"
      })
      require = jsonencode({
        data = "cold"
      })
    }

    downsample {
      fixed_interval = "1d"
      wait_timeout   = "12h"
    }
  }
}
