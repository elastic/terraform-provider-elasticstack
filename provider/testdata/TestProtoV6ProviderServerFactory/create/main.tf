provider "elasticstack" {
  elasticsearch {
    username  = "sup"
    password  = "dawg"
    endpoints = ["http://localhost:9200"]
  }
}
