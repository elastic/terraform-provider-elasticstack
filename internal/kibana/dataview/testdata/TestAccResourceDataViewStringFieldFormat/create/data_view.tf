variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_string_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "string_dv" {
  override = true
  data_view = {
    title           = "${var.index_name}*"
    name            = var.index_name
    time_field_name = "@timestamp"
    allow_no_index  = true
    field_formats = {
      host_name = {
        id = "string"
        params = {
          transform = "upper"
        }
      }
      message = {
        id = "truncate"
        params = {
          field_length = 20
        }
      }
    }
  }
}
