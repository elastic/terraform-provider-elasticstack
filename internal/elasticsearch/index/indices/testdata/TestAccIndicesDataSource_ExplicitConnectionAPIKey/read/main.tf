variable "index_name" {
  type = string
}

variable "endpoints" {
  type = list(string)
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "${var.index_name}-api-key"
  role_descriptors = jsonencode({
    indices_data_source = {
      cluster = ["monitor"]
      indices = [{
        names      = [var.index_name]
        privileges = ["read", "view_index_metadata"]
      }]
    }
  })
}

data "elasticstack_elasticsearch_indices" "test_conn" {
  target     = elasticstack_elasticsearch_index.test.name
  depends_on = [elasticstack_elasticsearch_index.test]

  elasticsearch_connection {
    endpoints = var.endpoints
    api_key   = elasticstack_elasticsearch_security_api_key.test.encoded
    headers = {
      XTerraformTest = "api-key"
      XTrace         = "indices"
    }
  }
}
