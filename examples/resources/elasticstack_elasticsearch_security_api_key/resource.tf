resource "elasticstack_elasticsearch_security_api_key" "api_key" {
  # Set the name
  name = "My API key"

  # Set the role descriptors
  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"],
      # The ES API expects `index`, however we use indices to be consistent with the roles API
      indices = [{
        names = ["index-a*"],
        privileges = ["read"]
      }]
    }
  })

  # Set the expiration for the API key
  expiration = "1d"

  # Set the custom metadata for this user
  metadata = jsonencode({
    "env"    = "testing"
    "open"   = false
    "number" = 49
  })
}

output "api_key" {
  value = elasticstack_elasticsearch_security_api_key.api_key
}
