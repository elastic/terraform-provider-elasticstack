provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "1d"
    }
  }

  delete {
    min_age = "30d"
    delete {}
  }
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "${var.policy_name}-api-key"
  role_descriptors = jsonencode({
    template_ilm_attachment = {
      cluster = ["manage_index_templates", "monitor"]
    }
  })
}

resource "elasticstack_elasticsearch_index_template_ilm_attachment" "test" {
  index_template = var.index_template
  lifecycle_name = elasticstack_elasticsearch_index_lifecycle.test.name

  elasticsearch_connection {
    endpoints = [var.endpoint]
    api_key   = elasticstack_elasticsearch_security_api_key.test.encoded
    headers = {
      XTerraformTest = "api-key"
    }
  }
}
