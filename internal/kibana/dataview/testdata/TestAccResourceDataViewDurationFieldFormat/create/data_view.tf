variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_duration_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "duration_dv" {
  override = true
  data_view = {
    title           = "${var.index_name}*"
    name            = var.index_name
    time_field_name = "@timestamp"
    allow_no_index  = true
    field_formats = {
      response_time = {
        id = "duration"
        params = {
          input_format              = "milliseconds"
          output_format             = "humanizePrecise"
          output_precision          = 2
          include_space_with_suffix = true
          use_short_suffix          = false
        }
      }
    }
  }
}
