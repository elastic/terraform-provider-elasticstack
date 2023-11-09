provider "elasticstack" {
  kibana {
    username  = "elastic"
    password  = "changeme"
    endpoints = ["http://localhost:5601"]
  }
}
