variable "name" {
  type = string
}

variable "endpoints" {
  type = list(string)
}

provider "elasticstack" {
  elasticsearch {}
}

# A role that can only manage API keys owned by its holder. monitor is needed
# for the cluster info lookups the provider makes; manage_api_key is not granted.
resource "elasticstack_elasticsearch_security_role" "own_keys_only" {
  name    = var.name
  cluster = ["manage_own_api_key", "monitor"]
}

resource "elasticstack_elasticsearch_security_user" "own_keys_only" {
  username = var.name
  password = "qwerty123"
  roles    = [elasticstack_elasticsearch_security_role.own_keys_only.name]
}

# Created, refreshed and invalidated by a user that holds only
# manage_own_api_key, with the default restrict_to_owned = false.
resource "elasticstack_elasticsearch_security_api_key" "default" {
  name = "${var.name}-default"

  elasticsearch_connection {
    endpoints = var.endpoints
    username  = elasticstack_elasticsearch_security_user.own_keys_only.username
    password  = elasticstack_elasticsearch_security_user.own_keys_only.password
    insecure  = true
  }
}

# Same, with restrict_to_owned = true.
resource "elasticstack_elasticsearch_security_api_key" "restricted" {
  name              = "${var.name}-restricted"
  restrict_to_owned = true

  elasticsearch_connection {
    endpoints = var.endpoints
    username  = elasticstack_elasticsearch_security_user.own_keys_only.username
    password  = elasticstack_elasticsearch_security_user.own_keys_only.password
    insecure  = true
  }
}
