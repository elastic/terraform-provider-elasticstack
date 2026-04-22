provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_cold" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }

  cold {
    min_age = "30d"

    set_priority {
      priority = 0
    }

    readonly {}

    allocate {
      number_of_replicas = 1
      include = jsonencode({
        box_type = "cold"
      })
    }
  }
}
