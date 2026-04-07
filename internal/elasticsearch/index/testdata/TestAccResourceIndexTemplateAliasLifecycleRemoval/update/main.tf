variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.name
  index_patterns = ["${var.name}-*"]

  data_stream {}

  template {
    alias {
      name = "detailed_alias_reset"
    }
  }
}
