resource "elasticstack_elasticsearch_index" "my_index" {
  name                = "my-index-000001"
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "my_data_view" {
  data_view = {
    title = "my-index-*"
    name  = "My Index Data View"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}

resource "elasticstack_kibana_default_data_view" "default" {
  data_view_id = elasticstack_kibana_data_view.my_data_view.data_view.id
  force        = true
  skip_delete  = false
}
