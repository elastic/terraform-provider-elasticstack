provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = "kibana-index-connector-example"
  mappings = jsonencode({
    properties = {
      alert_date = { type = "date", format = "date_optional_time||epoch_millis" }
    }
  })
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = elasticstack_elasticsearch_index.my_index.name
    executionTimeField = "alert_date"
  })
}
