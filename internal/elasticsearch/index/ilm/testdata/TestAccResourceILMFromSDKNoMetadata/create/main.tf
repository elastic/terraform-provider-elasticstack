provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "1d"
    }
  }

  warm {
    min_age = "7d"

    allocate {
      number_of_replicas = 1
    }
  }
}
