provider "elasticstack" {
  elasticsearch {}
}

variable "policy_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_toggles" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "7d"
    }

    readonly {
      enabled = false
    }

    unfollow {
      enabled = false
    }
  }

  warm {
    min_age = "14d"

    readonly {
      enabled = false
    }

    migrate {
      enabled = false
    }

    unfollow {
      enabled = false
    }
  }

  cold {
    min_age = "30d"

    readonly {
      enabled = false
    }

    migrate {
      enabled = false
    }

    freeze {
      enabled = false
    }

    unfollow {
      enabled = false
    }
  }
}
