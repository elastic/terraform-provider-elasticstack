variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false

  mappings = jsonencode({
    properties = {
      title = { type = "text" }
      body  = { type = "text" }
    }
  })

  lifecycle {
    ignore_changes = [settings_raw]
  }
}
