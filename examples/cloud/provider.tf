terraform {
  required_version = ">= 1.0.0"

  required_providers {
    ec = {
      source = "elastic/ec"
    }
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "~>0.7"
    }
  }
}

provider "ec" {
  # You can fill in your API key here, or use an environment variable TF_VAR_ec_apikey instead
  # For details on how to generate an API key, see: https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html.
  apikey = var.ec_apikey
}
