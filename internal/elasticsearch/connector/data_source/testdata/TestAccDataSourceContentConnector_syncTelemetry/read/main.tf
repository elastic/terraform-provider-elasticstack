variable "connector_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_connector" "lookup" {
  connector_id = var.connector_id
}
