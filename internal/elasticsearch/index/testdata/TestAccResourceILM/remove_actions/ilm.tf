provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = var.policy_name

  hot {
    min_age = "1h"

    set_priority {
      priority = 0
    }

    rollover {
      max_age = "2d"
    }
  }

  warm {
    min_age = "0ms"

    set_priority {
      priority = 60
    }

    readonly {}

    allocate {
      exclude = jsonencode({
        box_type = "hot"
      })
      number_of_replicas = 1
    }

    shrink {
      max_primary_shard_size = "50gb"
    }
  }
}
