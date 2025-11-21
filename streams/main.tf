terraform {
  required_providers {
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "0.12.2"
    }
  }
}
provider "elasticstack" {
  elasticsearch {
    endpoints = ["http://localhost:9200"]
    username  = var.elasticsearch_username
    password  = var.elasticsearch_password
  }

  kibana {
    endpoints = ["http://localhost:5601"] # adjust to your Kibana URL(remember about prefix)
    username  = var.kibana_username
    password  = var.kibana_password
  }
}

variable "elasticsearch_username" {
  type    = string
  default = "elastic"
}

variable "elasticsearch_password" {
  type    = string
  default = "changeme"
}

variable "kibana_username" {
  type    = string
  default = "elastic"
}
variable "kibana_password" {
  type    = string
  default = "changeme"
}

# resource "elasticstack_kibana_stream" "example_group" {
#   name        = "tf-example-group"
#   description = "Terraform example group stream"

#   group = {
#     members  = ["logs-synth.1-default", "logs-synth.2-default", "logs-synth.3-default"] # adjust to something valid in your env
#     metadata = { env = "dev" }
#     tags     = ["terraform", "poc"]
#   }
# }

resource "elasticstack_kibana_stream" "example_group" {
  name              = "tf-example-group-created-by-tf"
  create_if_missing = true
  group = {
    members  = ["logs-synth.1-default"]
    metadata = {}
    tags     = []
  }
}

resource "elasticstack_kibana_stream" "example_ingest" {
  name     = "logs-synth.1-default"
  space_id = "default"
}