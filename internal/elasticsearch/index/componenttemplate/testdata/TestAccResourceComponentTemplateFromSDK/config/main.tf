variable "template_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "upgrade" {
  name    = var.template_name
  version = 1

  metadata = jsonencode({
    description = "from sdk"
  })

  template {
    mappings = jsonencode({
      properties = { from_sdk = { type = "keyword" } }
    })
    settings = jsonencode({
      index = { number_of_shards = 1 }
    })

    alias {
      name = "my_alias"
    }
  }
}
