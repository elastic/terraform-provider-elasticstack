resource "elasticstack_elasticsearch_security_api_key" "api_key" {
  # Set the name
  name = "My API key"

  # Set the role descriptors
  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"],
      indices = [
        {
          names      = ["index-a*"],
          privileges = ["read"]
        }
      ]
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

# restriction on a role descriptor for an API key is supported since Elastic 8.9
resource "elasticstack_elasticsearch_security_api_key" "api_key_with_restriction" {
  # Set the name
  name = "My API key"
  # Set the role descriptors
  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"],
      indices = [
        {
          names      = ["index-a*"],
          privileges = ["read"]
        }
      ],
      restriction = {
        workflows = [ "search_application_query" ]
      }
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
  value     = elasticstack_elasticsearch_security_api_key.api_key
  sensitive = true
}
