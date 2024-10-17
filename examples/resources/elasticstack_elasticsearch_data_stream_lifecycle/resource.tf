provider "elasticstack" {
  elasticsearch {}
}

// First we must have a index template created
resource "elasticstack_elasticsearch_index_template" "my_data_stream_template" {
  name = "my_data_stream"

  index_patterns = ["my-stream*"]

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

// finally we can manage lifecycle of data stream
resource "elasticstack_elasticsearch_data_stream_lifecycle" "my_data_stream_lifecycle" {
  name           = "my-stream"
  data_retention = "3d"

  depends_on = [
    elasticstack_elasticsearch_data_stream.my_data_stream,
  ]
}

// or you can use wildcards to manage multiple lifecycles at once
resource "elasticstack_elasticsearch_data_stream_lifecycle" "my_data_stream_lifecycle_multiple" {
  name           = "stream-*"
  data_retention = "3d"
}
