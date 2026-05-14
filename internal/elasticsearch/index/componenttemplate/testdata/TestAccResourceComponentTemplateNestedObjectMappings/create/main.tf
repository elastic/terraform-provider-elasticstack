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
      properties = {
        event = {
          properties = {
            size = { type = "long" }
          }
        }
      }
    })
  }
}
