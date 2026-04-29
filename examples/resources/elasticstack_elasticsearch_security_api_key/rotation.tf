# Rotation uses the external hashicorp/time provider alongside elasticstack — install both when planning this example locally.

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

provider "time" {}

terraform {
  required_providers {
    elasticstack = {
      source = "elastic/elasticstack"
    }
    time = {
      source = "hashicorp/time"
    }
  }
}

resource "time_rotating" "api_key_rotation" {
  rotation_days = 1
}

resource "elasticstack_elasticsearch_security_api_key" "rotating_api_key" {
  name = "rotating-api-key-${time_rotating.api_key_rotation.id}"

  lifecycle {
    create_before_destroy = true
  }

  role_descriptors = jsonencode({
    rotating = {
      cluster = ["monitor"]
      indices = [
        {
          names      = ["logs-*"]
          privileges = ["read"]
        }
      ]
    }
  })
}
