variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_ilm" {
  name = var.name

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
    min_age = "2d"
    delete {}
  }
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name = var.name

  index_patterns = ["${var.name}*"]

  metadata = jsonencode({
    env     = "test"
    version = 1
  })

  template {
    settings = jsonencode({
      "lifecycle.name" = elasticstack_elasticsearch_index_lifecycle.test_ilm.name
    })
  }

  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "test_ds" {
  name = var.name

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}
