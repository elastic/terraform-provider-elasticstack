variable "index_name1" {
  description = "The name of the first Elasticsearch index"
  type        = string
}

variable "index_name2" {
  description = "The name of the second Elasticsearch index"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = var.index_name1
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "my_other_index" {
  name                = var.index_name2
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
  data_view = {
    title = "${var.index_name1}*"
  }
  depends_on = [elasticstack_elasticsearch_index.my_index]
}

resource "elasticstack_kibana_data_view" "dv2" {
  data_view = {
    title = "${var.index_name2}*"
  }
  depends_on = [elasticstack_elasticsearch_index.my_other_index]
}

resource "elasticstack_kibana_default_data_view" "test" {
  data_view_id = elasticstack_kibana_data_view.dv2.data_view.id
  force        = true
}