variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name = var.name

  index_patterns = ["${var.name}*"]

  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "test_ds_one" {
  name = "${var.name}-one"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}

resource "elasticstack_elasticsearch_data_stream" "test_ds_two" {
  name = "${var.name}-multiple-one"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}

resource "elasticstack_elasticsearch_data_stream" "test_ds_three" {
  name = "${var.name}-multiple-two"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}

resource "elasticstack_elasticsearch_data_stream_lifecycle" "test_ds_lifecycle" {
  name             = "${var.name}-one"
  enabled          = false
  expand_wildcards = "all"
  data_retention   = "2d"
  downsampling = [
    {
      after          = "2d"
      fixed_interval = "30m"
    },
    {
      after          = "9d"
      fixed_interval = "2d"
    }
  ]

  depends_on = [
    elasticstack_elasticsearch_data_stream.test_ds_one
  ]
}

resource "elasticstack_elasticsearch_data_stream_lifecycle" "test_ds_lifecycle_multiple" {
  name           = "${var.name}-multiple-*"
  data_retention = "2d"
  downsampling = [
    {
      after          = "1d"
      fixed_interval = "10m"
    },
    {
      after          = "7d"
      fixed_interval = "1d"
    }
  ]

  depends_on = [
    elasticstack_elasticsearch_data_stream.test_ds_two,
    elasticstack_elasticsearch_data_stream.test_ds_three
  ]
}
