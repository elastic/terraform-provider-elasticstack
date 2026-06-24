provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_delete_min_age" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }

  delete {
    min_age = "30d"

    delete {}
  }
}
