variable "index_name" {
  description = "The index name"
  type        = string
}

variable "space_id" {
  description = "The target Kibana space"
  type        = string
}

variable "data_view_id" {
  description = "The managed data view id"
  type        = string
}

variable "kibana_endpoint" {
  description = "The explicit Kibana endpoint for this test"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {
    endpoints = [var.kibana_endpoint]
  }
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_kibana_data_view" "dv" {
  space_id = elasticstack_kibana_space.test.space_id

  data_view = {
    id              = var.data_view_id
    title           = "${var.index_name}*"
    name            = var.index_name
    time_field_name = "@timestamp"
  }
}
