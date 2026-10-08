provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "issue_5019" {
  name = var.policy_name

  hot {
    min_age = "0ms"
    rollover {
      max_age = "90d"
    }
  }

  warm {
    min_age = "0ms"
    allocate {
      number_of_replicas    = 1
      total_shards_per_node = -1
      include               = jsonencode({})
      exclude               = jsonencode({})
    }
  }
}
