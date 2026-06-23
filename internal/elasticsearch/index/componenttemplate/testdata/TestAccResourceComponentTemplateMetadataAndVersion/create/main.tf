variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name     = var.name
  metadata = jsonencode({ env = "prod" })
  version  = 1

  template {}
}
