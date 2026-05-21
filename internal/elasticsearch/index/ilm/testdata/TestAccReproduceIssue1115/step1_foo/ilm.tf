provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "foo" {
  name = var.policy_name

  hot {
    set_priority {
      priority = 100
    }
  }
}
