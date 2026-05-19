variable "index_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "idx" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "fa_dv" {
  data_view = {
    title           = "${var.index_name}*"
    name            = var.index_name
    time_field_name = "@timestamp"
    field_attrs = {
      "host.hostname" = { custom_label = "Host" }
    }
  }
  depends_on = [elasticstack_elasticsearch_index.idx]
}
