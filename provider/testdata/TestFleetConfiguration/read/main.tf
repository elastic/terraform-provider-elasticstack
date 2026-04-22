variable "fleet_endpoint" {
  type = string
}

variable "fleet_username" {
  type = string
}

variable "fleet_password" {
  type = string
}

variable "fleet_ca_certs" {
  type    = list(string)
  default = []
}

provider "elasticstack" {
  fleet {
    endpoint = var.fleet_endpoint
    username = var.fleet_username
    password = var.fleet_password
    ca_certs = var.fleet_ca_certs
  }
}

data "elasticstack_fleet_enrollment_tokens" "test" {}
