# Creating a deployment on Elastic Cloud GCP region,
# with elasticsearch and kibana components.
resource "ec_deployment" "cluster" {
  region                 = "gcp-us-central1"
  name			         = "mydeployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = "gcp-storage-optimized"

  elasticsearch {}

  kibana {}
}

# Fetching the latest version of the stack availalbe on Elastic Cloud
data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "gcp-us-central1"
}

# Defining a user for ingesting
resource "elasticstack_elasticsearch_security_user" "user" {
  username = "ingest_user"

  # Password is cleartext here for comfort, but there's also a hashed password option
  password = "mysecretpassword"
  roles    = ["editor"]

  # Set the custom metadata for this user
  metadata = jsonencode({
    "env"    = "testing"
    "open"   = false
    "number" = 49
  })

  # Use our Elastic Cloud deployemnt outputs for connection details.
  # This also allows the provider to create the proper relationships between the two resources.
  elasticsearch_connection {
    endpoints = ["${ec_deployment.cluster.elasticsearch[0].https_endpoint}"]
    username  = "${ec_deployment.cluster.elasticsearch_username}"
    password  = "${ec_deployment.cluster.elasticsearch_password}"
  }
}

# Configuring my cluster with an index template as well.
resource "elasticstack_elasticsearch_index_template" "my_template" {
  name = "my_ingest_1"

  priority = 42
  index_patterns = ["server-logs*"]

  template {
    aliases {
      name = "my_template_test"
    }

      settings = jsonencode({
        number_of_shards = "3"
      })

      mappings = jsonencode({
        properties: {
          "@timestamp": { "type": "date" },
          "username": {"type": "keyword" }
        }
      })
  }

  elasticsearch_connection {
    endpoints = ["${ec_deployment.cluster.elasticsearch[0].https_endpoint}"]
    username  = "${ec_deployment.cluster.elasticsearch_username}"
    password  = "${ec_deployment.cluster.elasticsearch_password}"
  }
}
