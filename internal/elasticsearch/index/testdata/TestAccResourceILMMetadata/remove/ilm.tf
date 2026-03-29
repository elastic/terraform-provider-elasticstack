provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_meta" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }
}
