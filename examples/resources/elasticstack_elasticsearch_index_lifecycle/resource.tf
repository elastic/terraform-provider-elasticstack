provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "my_ilm" {
  name = "my_ilm_policy"

  hot {
    min_age = "1h"
    set_priority {
      priority = 0
    }
    rollover {
      max_age = "1d"
    }
    readonly {}
  }

  warm {
    min_age = "0ms"
    set_priority {
      priority = 10
    }
    readonly {}
    allocate {
      exclude = jsonencode({
        box_type = "hot"
      })
      number_of_replicas    = 1
      total_shards_per_node = 200
    }
  }

  delete {
    min_age = "2d"
    delete {}
  }
}
