variable "prefix" {
  type = string
}

variable "rollover_alias" {
  type = string
}

variable "policy_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "30d"
    }
  }

  delete {
    min_age = "90d"
    delete {}
  }
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "${var.prefix}-template"
  index_patterns = ["${var.prefix}-*"]
  priority       = 500

  template {
    settings = jsonencode({
      number_of_shards   = 1
      number_of_replicas = 0
      refresh_interval   = "30s"
      index = {
        lifecycle = {
          name           = elasticstack_elasticsearch_index_lifecycle.test.name
          rollover_alias = var.rollover_alias
        }
      }
    })
    mappings = jsonencode({
      properties = {
        order_id = { type = "keyword" }
      }
    })
  }
}
