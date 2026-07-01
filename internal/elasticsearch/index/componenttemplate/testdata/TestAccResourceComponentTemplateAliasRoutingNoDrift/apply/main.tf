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
      name    = "routing_only_alias"
      routing = "shard_1"
    }
  }
}
