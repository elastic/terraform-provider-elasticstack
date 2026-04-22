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
      number_of_replicas = 0
    }

    downsample {
      fixed_interval = "2d"
    }
  }
}
