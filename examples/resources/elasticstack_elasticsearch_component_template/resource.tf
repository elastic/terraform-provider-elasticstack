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

# Enable the failure store on data streams composed from this component.
# Requires Elasticsearch >= 9.1.0.
resource "elasticstack_elasticsearch_component_template" "failure_store_custom" {
  name = "logs-myapp@custom"

  template {
    data_stream_options {
      failure_store {
        enabled = true
        lifecycle {
          data_retention = "30d"
        }
      }
    }
  }
}
