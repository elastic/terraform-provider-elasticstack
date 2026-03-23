provider "elasticstack" {
  elasticsearch {
    bearer_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    endpoints    = ["http://localhost:9200"]
  }
}
