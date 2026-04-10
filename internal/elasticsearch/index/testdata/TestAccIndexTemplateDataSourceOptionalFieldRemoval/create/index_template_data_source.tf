provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]
  version        = 7
  metadata       = jsonencode({ owner = "team-a", description = "initial" })

  data_stream {}

  template {
    alias {
      name = "removal_alias"
    }

    mappings = jsonencode({
      properties = {
        log_level = {
          type = "keyword"
        }
      }
    })
    settings = jsonencode({
      index = {
        number_of_shards = "1"
      }
    })

    lifecycle {
      data_retention = "30d"
    }
  }
}

data "elasticstack_elasticsearch_index_template" "test" {
  name = elasticstack_elasticsearch_index_template.test.name
}
