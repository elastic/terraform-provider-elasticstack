variable "endpoints" {
  type = string
}

variable "bearer_token" {
  type = string
}

provider "elasticstack" {
  elasticsearch {
    endpoints    = [var.endpoints]
    bearer_token = var.bearer_token
  }
}

data "elasticstack_elasticsearch_security_user" "test" {
  username = "elastic"
}
