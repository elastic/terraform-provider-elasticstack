variable "api_key_name" {
  type = string
}

variable "epoch" {
  type = string
}

terraform {
  required_providers {
    time = {
      source = "hashicorp/time"
    }
  }
}

provider "elasticstack" {
  elasticsearch {}
}

provider "time" {}

resource "time_rotating" "api_key_rotation" {
  rotation_minutes = 1
  triggers = {
    epoch = var.epoch
  }
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "${var.api_key_name}-${time_rotating.api_key_rotation.id}"

  lifecycle {
    create_before_destroy = true
  }

  role_descriptors = jsonencode({
    rotate-test = {
      cluster = ["monitor"]
      indices = [{
        names                    = ["logs-*"]
        privileges               = ["read"]
        allow_restricted_indices = false
      }]
    }
  })

  expiration = "1d"
}
