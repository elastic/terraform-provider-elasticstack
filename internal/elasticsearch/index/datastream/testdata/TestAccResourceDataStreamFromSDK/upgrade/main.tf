variable "name" {
  description = "The data stream name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name           = var.name
  index_patterns = ["${var.name}*"]
  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "test_ds" {
  name = var.name

  depends_on = [elasticstack_elasticsearch_index_template.test_ds_template]
}
