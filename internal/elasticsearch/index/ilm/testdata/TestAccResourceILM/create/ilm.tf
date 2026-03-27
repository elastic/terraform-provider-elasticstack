provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = var.policy_name

  hot {
    min_age = "1h"

    set_priority {
      priority = 10
    }

    rollover {
      max_age = "1d"
    }

    readonly {}
  }

  delete {
    delete {}
  }
}
