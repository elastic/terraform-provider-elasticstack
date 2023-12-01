terraform {
  required_version = ">= 1.0.0"
  required_providers {
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "~>0.9"
    }
  }
}

provider "elasticstack" {
  elasticsearch {
    username  = "elastic"
    password  = "changeme"
    endpoints = ["http://localhost:9200"]
  }
}
