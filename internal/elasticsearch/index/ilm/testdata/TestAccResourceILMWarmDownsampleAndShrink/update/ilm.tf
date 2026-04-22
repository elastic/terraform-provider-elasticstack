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
      number_of_replicas = 1
      exclude = jsonencode({
        box_type = "hot"
      })
    }

    downsample {
      fixed_interval = "2d"
    }

    shrink {
      max_primary_shard_size = "50gb"
    }
  }
}
