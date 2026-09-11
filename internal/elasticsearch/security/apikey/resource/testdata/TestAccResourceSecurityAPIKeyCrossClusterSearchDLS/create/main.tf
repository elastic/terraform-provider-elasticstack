variable "api_key_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = var.api_key_name
  type = "cross_cluster"

  access = {
    search = [
      {
        names          = ["logs-*"]
        field_security = jsonencode({ grant = ["field1", "field2"] })
        query          = jsonencode({ match = { field1 = "value1" } })
      }
    ]
  }
}
