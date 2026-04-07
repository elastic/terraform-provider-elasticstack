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
        names = ["log-*", "metrics-*"]
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
    description = "Cross-cluster test key"
    environment = "test"
  })
}
