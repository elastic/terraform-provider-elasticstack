variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
  data_view = {
    title = "${var.index_name}*"
  }
}
