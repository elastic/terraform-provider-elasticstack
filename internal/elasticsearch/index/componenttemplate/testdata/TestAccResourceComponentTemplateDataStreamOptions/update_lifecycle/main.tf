variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name = var.name

  template {
    data_stream_options {
      failure_store {
        enabled = true
        lifecycle {
          data_retention = "14d"
        }
      }
    }
  }
}