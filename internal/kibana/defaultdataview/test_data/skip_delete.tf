variable "index_name" {
  description = "The name of the Elasticsearch index"
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
  depends_on = [elasticstack_elasticsearch_index.my_index]
}

resource "elasticstack_kibana_default_data_view" "test" {
  data_view_id = elasticstack_kibana_data_view.dv.data_view.id
  force        = true
  skip_delete  = true
}