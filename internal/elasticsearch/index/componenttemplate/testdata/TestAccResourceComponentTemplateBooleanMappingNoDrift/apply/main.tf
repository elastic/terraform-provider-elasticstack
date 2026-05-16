variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name = var.name

  template {
    mappings = jsonencode({
      date_detection    = true
      dynamic           = false
      numeric_detection = true
    })
  }
}
