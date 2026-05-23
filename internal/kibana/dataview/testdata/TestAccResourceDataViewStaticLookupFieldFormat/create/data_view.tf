variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_lookup_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "lookup_dv" {
  override = true
  data_view = {
    title           = "${var.index_name}*"
    name            = var.index_name
    time_field_name = "@timestamp"
    allow_no_index  = true
    field_formats = {
      status_code = {
        id = "static_lookup"
        params = {
          lookup_entries = [
            { key = "200", value = "OK" },
            { key = "404", value = "Not Found" },
          ]
          unknown_key_value = "Unknown"
        }
      }
    }
  }
}
