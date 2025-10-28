variable "index_name" {
  description = "The name of the Elasticsearch index"
  type        = string
}

variable "space_id" {
  description = "The ID of the Kibana space"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = var.space_id
  name        = "Test Space ${var.space_id}"
  description = "Test space for default data view"
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
  space_id = elasticstack_kibana_space.test_space.space_id
  data_view = {
    title = "${var.index_name}*"
  }
  depends_on = [elasticstack_elasticsearch_index.my_index]
}

resource "elasticstack_kibana_default_data_view" "test" {
  space_id     = elasticstack_kibana_space.test_space.space_id
  data_view_id = elasticstack_kibana_data_view.dv.data_view.id
  force        = true
}