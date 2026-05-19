variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name = var.name

  template {
    settings = jsonencode({
      index = {
        number_of_shards = "1"
        routing = {
          allocation = {
            include = {
              _tier_preference = null
            }
          }
        }
      }
    })
  }
}
