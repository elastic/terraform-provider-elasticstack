provider "elasticstack" {
  elasticsearch {
    username  = "elastic"
    password  = "changeme"
    endpoints = ["http://localhost:9200"]
  }
}
