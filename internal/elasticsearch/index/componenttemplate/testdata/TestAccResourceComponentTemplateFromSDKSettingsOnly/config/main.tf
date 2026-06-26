variable "template_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "upgrade" {
  name    = var.template_name
  version = 1

  template {
    settings = jsonencode({
      index = { number_of_replicas = 1 }
    })
  }
}
