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

  warm {
    min_age = "0ms"
    shrink {
      number_of_shards = 1
    }
  }
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = "${var.prefix}-template"
  index_patterns = ["${var.prefix}-*"]
  priority       = 500

  template {
    settings = jsonencode({
      number_of_shards   = 2
      number_of_replicas = 0
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

resource "elasticstack_elasticsearch_rollover_index_series" "test" {
  prefix              = var.prefix
  rollover_alias      = var.rollover_alias
  deletion_protection = false

  depends_on = [elasticstack_elasticsearch_index_template.test]
}
