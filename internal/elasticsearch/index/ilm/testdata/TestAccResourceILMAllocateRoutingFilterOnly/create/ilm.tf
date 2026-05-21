provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_allocate_filter_only" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }

  warm {
    min_age = "7d"

    allocate {
      require = jsonencode({
        zone = "zone-1"
      })
    }
  }
}
