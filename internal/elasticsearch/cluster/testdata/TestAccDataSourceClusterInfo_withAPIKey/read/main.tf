variable "endpoint" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "cluster-info-data-source-api-key"
  role_descriptors = jsonencode({
    cluster_info = {
      cluster = ["monitor"]
    }
  })
}

data "elasticstack_elasticsearch_info" "test_conn" {
  elasticsearch_connection {
    endpoints = [var.endpoint]
    api_key   = elasticstack_elasticsearch_security_api_key.test.encoded
  }
}
