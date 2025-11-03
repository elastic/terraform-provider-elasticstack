variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
  override = true
  data_view = {
    title           = "${var.index_name}*"
    name            = var.index_name
    time_field_name = "@timestamp"
    source_filters  = ["event_time", "machine.ram"]
    allow_no_index  = true
    namespaces      = ["default", "foo", "bar"]
    field_formats = {
      event_time = {
        id = "date_nanos"
      }
      "machine.ram" = {
        id = "number"
        params = {
          pattern = "0,0.[000] b"
        }
      }
    }
    runtime_field_map = {
      runtime_shape_name = {
        type          = "keyword"
        script_source = "emit(doc['shape_name'].value)"
      }
    }
    field_attrs = {
      ingest_failure = { custom_label = "error.ingest_failure", count = 6 }
    }
  }
}
