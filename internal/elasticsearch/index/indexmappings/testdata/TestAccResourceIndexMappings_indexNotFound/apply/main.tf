variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_mappings" "test" {
  index = var.index_name

  mappings = jsonencode({
    properties = {
      title = { type = "text" }
    }
  })
}
