variable "connector_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_connector" "test" {
  connector_id = var.connector_id
  service_type = "postgresql"
  name         = "TF acc ds native connector"
  description  = "data source native connector acceptance test"
  index_name   = "content-connector-${var.connector_id}"
  is_native    = true
}

data "elasticstack_elasticsearch_connector" "lookup" {
  connector_id = elasticstack_elasticsearch_connector.test.connector_id
}
