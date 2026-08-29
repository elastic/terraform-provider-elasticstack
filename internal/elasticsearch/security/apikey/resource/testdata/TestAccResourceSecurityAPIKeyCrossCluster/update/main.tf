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
        names                    = ["log-*", "metrics-*"]
        allow_restricted_indices = false
        field_security           = jsonencode({ grant = ["field1", "field2", "field3"] })
        query                    = jsonencode({ match = { field1 = "value2" } })
      }
    ]
    replication = [
      {
        names = ["archives-*"]
      }
    ]
  }

  expiration = "30d"

  metadata = jsonencode({
    description = "Cross-cluster test key updated"
    environment = "test"
  })
}
