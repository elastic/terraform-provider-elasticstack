// Create an ILM policy for our data stream
resource "elasticstack_elasticsearch_index_lifecycle" "my_lifecycle_policy" {
  name = "my_lifecycle_policy"

  hot {
    min_age = "1h"
    set_priority {
      priority = 10
    }
    rollover {
      max_age = "1d"
    }
    readonly {}
  }

  delete {
    min_age = "2d"
    delete {}
  }
}

// Create a component template for mappings
resource "elasticstack_elasticsearch_component_template" "my_mappings" {
  name = "my_mappings"
  template {
    mappings = jsonencode({
      properties = {
        field1       = { type = "keyword" }
        field2       = { type = "text" }
        "@timestamp" = { type = "date" }
      }
    })
  }
}

// Create a component template for index settings
resource "elasticstack_elasticsearch_component_template" "my_settings" {
  name = "my_settings"
  template {
    settings = jsonencode({
      "lifecycle.name" = elasticstack_elasticsearch_index_lifecycle.my_lifecycle_policy.name
    })
  }
}

// Create an index template that uses the component templates
resource "elasticstack_elasticsearch_index_template" "my_index_template" {
  name           = "my_index_template"
  priority       = 500
  index_patterns = ["my-data-stream*"]
  composed_of    = [elasticstack_elasticsearch_component_template.my_mappings.name, elasticstack_elasticsearch_component_template.my_settings.name]
  data_stream {}
}

// Create a data stream based on the index template
resource "elasticstack_elasticsearch_data_stream" "my_data_stream" {
  name = "my-data-stream"

  // Make sure that template is created before the data stream
  depends_on = [
    elasticstack_elasticsearch_index_template.my_index_template
  ]
}
