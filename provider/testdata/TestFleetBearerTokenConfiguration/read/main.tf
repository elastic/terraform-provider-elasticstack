variable "fleet_endpoint" {
  type = string
}

variable "bearer_token" {
  type = string
}

variable "fleet_ca_certs" {
  type    = list(string)
  default = []
}

provider "elasticstack" {
  fleet {
    endpoint     = var.fleet_endpoint
    bearer_token = var.bearer_token
    ca_certs     = var.fleet_ca_certs
  }
}

data "elasticstack_fleet_enrollment_tokens" "test" {}
