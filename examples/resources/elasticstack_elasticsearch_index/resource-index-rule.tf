resource "elasticstack_elasticsearch_index" "my_index" {
  name = "my-index"
  mappings = jsonencode({
    properties = {
      alert_date = { type = "date", format = "date_optional_time||epoch_millis" }
      rule_id    = { type = "text" }
      rule_name  = { type = "text" }
      message    = { type = "text" }
    }
  })
}