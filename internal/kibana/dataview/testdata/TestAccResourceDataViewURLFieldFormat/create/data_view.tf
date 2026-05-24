variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_url_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "url_dv" {
  override = true
  data_view = {
    title           = "${var.index_name}*"
    name            = var.index_name
    time_field_name = "@timestamp"
    allow_no_index  = true
    field_formats = {
      thumbnail = {
        id = "url"
        params = {
          type          = "img"
          urltemplate   = "https://example.com/images/{{value}}"
          labeltemplate = "Image: {{value}}"
          width         = 200
          height        = 150
        }
      }
    }
  }
}
