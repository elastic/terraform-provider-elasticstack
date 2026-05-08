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
      name = "my_template_test"
    }

    settings = jsonencode({
      index = { number_of_shards = "3" }
    })
  }
}
