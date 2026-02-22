variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_settings_removing_field" {
  name = var.index_name

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
      field2 = { type = "text" }
    }
  })

  lifecycle {
    prevent_destroy = true
  }
}
