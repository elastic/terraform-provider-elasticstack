resource "elasticstack_elasticsearch_index_template" "my_template" {
  name = "my_template"

  priority       = 42
  index_patterns = ["logstash*", "filebeat*"]

  template {
    aliases {
      name = "my_template_test"
    }

    settings = jsonencode({
      number_of_shards = "3"
    })
  }
}
