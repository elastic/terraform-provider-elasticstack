provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]

  template {
    alias {
      name = "empty_template_alias"
    }

    settings = jsonencode({
      index = {
        number_of_shards = "2"
      }
    })
  }
}
