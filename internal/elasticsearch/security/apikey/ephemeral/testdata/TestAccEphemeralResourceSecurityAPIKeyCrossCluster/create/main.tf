variable "api_key_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

ephemeral "elasticstack_elasticsearch_security_api_key" "test" {
  name = var.api_key_name
  type = "cross_cluster"

  access = {
    search = [
      {
        names                    = ["logs-*", "metrics-*"]
        allow_restricted_indices = true
      }
    ]
    replication = [
      {
        names = ["archive-*"]
      }
    ]
  }

  expiration = "30d"

  metadata = jsonencode({
    description = "Cross-cluster ephemeral test key"
    environment = "test"
  })

  invalidate_on_close = false
}

provider "echo" {
  data = ephemeral.elasticstack_elasticsearch_security_api_key.test
}

resource "echo" "capture" {}
