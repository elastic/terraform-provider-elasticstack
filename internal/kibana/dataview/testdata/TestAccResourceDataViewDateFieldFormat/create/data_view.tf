variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_date_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "date_dv" {
  override = true
  data_view = {
    title           = "${var.index_name}*"
    name            = var.index_name
    time_field_name = "@timestamp"
    allow_no_index  = true
    field_formats = {
      last_login = {
        id = "date"
        params = {
          pattern  = "MMM D, YYYY @ HH:mm:ss.SSS"
          timezone = "America/New_York"
        }
      }
    }
  }
}
