provider "elasticstack" {
  elasticsearch {}
}

variable "template_name" {
  type = string
}

variable "endpoints" {
  type = list(string)
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.template_name
  index_patterns = ["${var.template_name}-*"]
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "${var.template_name}-api-key"
  role_descriptors = jsonencode({
    index_template_data_source = {
      cluster = ["manage_index_templates", "monitor"]
    }
  })
}

data "elasticstack_elasticsearch_index_template" "test_conn" {
  name = elasticstack_elasticsearch_index_template.test.name

  elasticsearch_connection {
    endpoints = var.endpoints
    api_key   = elasticstack_elasticsearch_security_api_key.test.encoded
    headers = {
      XTerraformTest = "api-key"
      XTrace         = "index-template"
    }
    insecure = false
  }
}
