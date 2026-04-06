variable "kibana_endpoint" {
  type = string
}

variable "kibana_bearer_token" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {
    endpoints    = [var.kibana_endpoint]
    bearer_token = var.kibana_bearer_token
  }
}

resource "elasticstack_kibana_space" "acc_test" {
  space_id = "acc_test_space"
  name     = "Acceptance Test Space"
}
