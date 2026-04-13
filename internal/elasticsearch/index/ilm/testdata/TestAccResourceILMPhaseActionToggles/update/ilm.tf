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
      enabled = true
    }

    unfollow {
      enabled = true
    }
  }

  warm {
    min_age = "14d"

    readonly {
      enabled = true
    }

    migrate {
      enabled = true
    }

    unfollow {
      enabled = true
    }
  }

  cold {
    min_age = "30d"

    readonly {
      enabled = true
    }

    migrate {
      enabled = true
    }

    freeze {
      enabled = true
    }

    unfollow {
      enabled = true
    }
  }
}
