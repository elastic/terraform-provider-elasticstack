variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_color_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "color_dv" {
  override = true
  data_view = {
    title           = "${var.index_name}*"
    name            = var.index_name
    time_field_name = "@timestamp"
    allow_no_index  = true
    field_formats = {
      status = {
        id = "color"
        params = {
          field_type = "string"
          colors = [
            {
              range      = "-Infinity:Infinity"
              regex      = "Completed"
              text       = "#000000"
              background = "#54B399"
            },
            {
              range      = "-Infinity:Infinity"
              regex      = "Error"
              text       = "#FFFFFF"
              background = "#BD271E"
            }
          ]
        }
      }
    }
  }
}
