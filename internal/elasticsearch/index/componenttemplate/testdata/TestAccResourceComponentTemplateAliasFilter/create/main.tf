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
      name   = "filtered_alias"
      filter = jsonencode({ term = { status = "active" } })
    }
  }
}
