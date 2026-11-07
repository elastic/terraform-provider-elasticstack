variable "kibana_endpoint" {
  type = string
}

variable "fleet_endpoint" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {
    endpoints = [var.kibana_endpoint]
    headers = {
      "cf-access-token" = "xyz"
    }
  }
  fleet {
    endpoint = var.fleet_endpoint
  }
}

data "elasticstack_fleet_enrollment_tokens" "test" {}
