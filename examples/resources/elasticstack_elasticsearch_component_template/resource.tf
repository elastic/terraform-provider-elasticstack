provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "my_template" {
  name = "my_template"

  template {
    alias {
      name = "my_template_test"
    }

    settings = jsonencode({
      number_of_shards = "3"
    })
  }
}

resource "elasticstack_elasticsearch_index_template" "my_template" {
  name = "my_data_stream"

  index_patterns = ["stream*"]
  composed_of    = [elasticstack_elasticsearch_component_template.my_template.name]
}
