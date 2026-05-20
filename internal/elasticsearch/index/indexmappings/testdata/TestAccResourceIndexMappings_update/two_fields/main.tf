variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false
  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_mappings" "test" {
  index = elasticstack_elasticsearch_index.test.name

  mappings = jsonencode({
    properties = {
      title = { type = "text" }
      body  = { type = "text" }
    }
  })
}
