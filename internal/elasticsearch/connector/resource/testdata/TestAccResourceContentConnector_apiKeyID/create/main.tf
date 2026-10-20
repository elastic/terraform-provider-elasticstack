variable "connector_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_connector" "test" {
  connector_id      = var.connector_id
  service_type      = "postgresql"
  name              = "TF acc api key"
  api_key_id        = "tf-acc-api-key-id-1"
  api_key_secret_id = "tf-acc-api-key-secret-id-1"
}
