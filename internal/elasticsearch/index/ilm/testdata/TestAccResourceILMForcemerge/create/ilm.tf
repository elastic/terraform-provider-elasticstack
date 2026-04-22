provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_forcemerge" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }
  }

  warm {
    min_age = "7d"

    forcemerge {
      max_num_segments = 1
      index_codec      = "best_compression"
    }
  }
}
