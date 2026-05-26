variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name = var.name

  template {
    alias {
      name = "explicit_empty_object_test"
    }

    mappings = jsonencode({})
    settings = jsonencode({})
  }
}
