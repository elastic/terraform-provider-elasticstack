variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "ds_name" {
  description = "The data stream name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name           = var.ds_name
  index_patterns = [var.ds_name]
  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "test_ds" {
  name = var.ds_name
  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}

resource "elasticstack_elasticsearch_index_alias" "test_alias" {
  name = var.alias_name

  write_index = {
    name = elasticstack_elasticsearch_data_stream.test_ds.name
  }
}
