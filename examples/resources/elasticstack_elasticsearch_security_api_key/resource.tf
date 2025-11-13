provider "elasticstack" {
  elasticsearch {}
}

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
        workflows = ["search_application_query"]
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

# Example: Cross-cluster API key
resource "elasticstack_elasticsearch_security_api_key" "cross_cluster_key" {
  name = "My Cross-Cluster API Key"
  type = "cross_cluster"

  # Define access permissions for cross-cluster operations
  access = {

    # Grant replication access to specific indices  
    replication = [
      {
        names = ["archive-*"]
      }
    ]
  }

  # Set the expiration for the API key
  expiration = "30d"

  # Set arbitrary metadata
  metadata = jsonencode({
    description = "Cross-cluster key for production environment"
    environment = "production"
    team        = "platform"
  })
}

output "cross_cluster_api_key" {
  value     = elasticstack_elasticsearch_security_api_key.cross_cluster_key
  sensitive = true
}
