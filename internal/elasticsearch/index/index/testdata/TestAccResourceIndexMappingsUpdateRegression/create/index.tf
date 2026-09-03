variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name = var.index_name

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection = false
}
