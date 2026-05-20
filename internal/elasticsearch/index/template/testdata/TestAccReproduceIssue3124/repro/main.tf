variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name           = var.name
  index_patterns = ["${var.name}-*"]

  template {
    settings = jsonencode({
      index = {
        number_of_shards   = "1"
        number_of_replicas = "0"
        search = {
          slowlog = {
            include = {
              user = "true"
            }
            threshold = {
              query = {
                warn = "10s"
              }
            }
          }
        }
        lifecycle = {
          parse_origination_date = "true"
        }
      }
    })
  }
}
