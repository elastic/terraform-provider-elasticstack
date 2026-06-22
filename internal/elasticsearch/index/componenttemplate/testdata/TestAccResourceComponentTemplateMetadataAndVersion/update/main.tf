variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name     = var.name
  metadata = jsonencode({ env = "staging" })
  version  = 2

  template {}
}
