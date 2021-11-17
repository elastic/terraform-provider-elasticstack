terraform {
  required_version = ">= 1.0.0"

  required_providers {
    elasticstack = {
      source = "elastic/elasticstack"
      version = "~>0.1.0"
    }
  }
}

provider "elasticstack" {
  elasticsearch {
    endpoints = ["http://localhost:9200"]
    # You can also use authentication if needed
    # username = "elastic"
    # password = "changeme"
  }
}
