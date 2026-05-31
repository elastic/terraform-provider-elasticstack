variable "connector_id" {
  type = string
}

variable "secret_value" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_connector" "test" {
  connector_id = var.connector_id
  service_type = "postgresql"
  name         = "TF acc config secret"

  scheduling = {
    full = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
    incremental = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
    access_control = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
  }


  configuration_values = {
    password = { secret_value = var.secret_value }
  }
}
