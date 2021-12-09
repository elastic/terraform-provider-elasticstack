provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "user" {
  username = "testuser"

  // use hashed password: see https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-user.html#security-api-put-user-request-body
  password_hash = "$2a$10$rMZe6TdsUwBX/TA8vRDz0OLwKAZeCzXM4jT3tfCjpSTB8HoFuq8xO"
  roles         = ["kibana_user"]

  // set the custom metadata for this user
  metadata = jsonencode({
    "env"    = "testing"
    "open"   = false
    "number" = 49
  })

  elasticsearch_connection {
    endpoints = ["http://localhost:9200"]
    username  = "elastic"
    password  = "changeme"
  }
}

resource "elasticstack_elasticsearch_security_user" "dev" {
  username = "devuser"

  password = "1234567890"
  roles    = ["kibana_user"]

  // set the custom metadata for this user
  metadata = jsonencode({
    "env"    = "testing"
    "open"   = false
    "number" = 49
  })
}
