provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_coverage" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }

  warm {
    min_age = "1d"

    migrate {
      enabled = false
    }

    allocate {
      require = jsonencode({
        box_type = "warm"
      })
    }

    shrink {
      number_of_shards = 1
    }
  }
}
