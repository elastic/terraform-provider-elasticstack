terraform {
  required_providers {
    elasticstack = {
      source = "elastic/elasticstack"
    }
  }
}

provider "elasticstack" {
  elasticsearch {
    username  = "elastic"
    password  = "password"
    endpoints = ["http://localhost:9200"]
  }
  kibana {
    username  = "elastic"
    password  = "password"
    endpoints = ["http://localhost:5601"]
  }
}

resource "elasticstack_kibana_action_connector" "slack-connector" {
  name              = "slack"
  connector_type_id = ".slack"
  secrets = jsonencode({
    webhookUrl = "https://lol.com"
  })
}

data "elasticstack_kibana_action_connector" "example" {
  name              = "myslackconnector"
  space_id          = "default"
  connector_type_id = ".slack"
  depends_on        = [elasticstack_kibana_action_connector.slack-connector]
}

output "connector_id" {
  # value = elasticstack_kibana_action_connector.slack-connector.connector_id
  value = data.elasticstack_kibana_action_connector.example.connector_id
}
