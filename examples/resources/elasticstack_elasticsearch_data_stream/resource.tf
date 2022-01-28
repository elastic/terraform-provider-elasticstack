provider "elasticstack" {
  elasticsearch {}
}

// Create an ILM policy for our data stream
resource "elasticstack_elasticsearch_index_lifecycle" "my_ilm" {
  name = "my_ilm_policy"

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

// First we must have a index template created
resource "elasticstack_elasticsearch_index_template" "my_data_stream_template" {
  name = "my_data_stream"

  index_patterns = ["my-stream*"]

  template {
    // make sure our template uses prepared ILM policy
    settings = jsonencode({
      "lifecycle.name" = elasticstack_elasticsearch_index_lifecycle.my_ilm.name
    })
  }

  data_stream {}
}

// and now we can create data stream based on the index template
resource "elasticstack_elasticsearch_data_stream" "my_data_stream" {
  name = "my-stream"

  // make sure that template is created before the data stream
  depends_on = [
    elasticstack_elasticsearch_index_template.my_data_stream_template
  ]
}
