provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

variable "metadata" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_meta" {
  name     = var.policy_name
  metadata = var.metadata

  hot {
    rollover {
      max_age = "7d"
    }
  }
}
