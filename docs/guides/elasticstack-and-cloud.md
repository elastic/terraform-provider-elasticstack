---
subcategory: ""
page_title: "Using the Elastic Stack provider with Elastic Cloud"
description: |-
    An example of how to spin up a deployment and configure it in a single plan.
---

# Using the Elastic Stack provider with Elastic Cloud



## Creating deployments

A common scenario for using the Elastic Stack provider, is to manage & configure Elastic Cloud deployments.
In order to do that, we'll use both the Elastic Cloud provider, as well as the Elastic Stack provider.

Start off by configuring just the Elastic Cloud provider in a `provider.tf` file for example:

```terraform
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    ec = {
      source = "elastic/ec"
    }
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "0.12.3"
    }
  }
}

provider "ec" {
  # You can fill in your API key here, or use an environment variable TF_VAR_ec_apikey instead
  # For details on how to generate an API key, see: https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html.
  apikey = var.ec_apikey
}
```

Note that the provider needs to be configured with an API key which has to be created in the [Elastic Cloud console](https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html). With the provider configured, we can now use it to create some deployments. Here, we're creating two deployments -- one for your data, and the other is setup as a separate monitor. 

```terraform
# Creating deployments on Elastic Cloud GCP region with elasticsearch and kibana components. One deployment is a dedicated monitor for the other. 

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = var.region
}

resource "ec_deployment" "monitoring" {
  region                 = var.region
  name                   = "my-monitoring-deployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = var.deployment_template_id

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

resource "ec_deployment" "cluster" {
  region                 = var.region
  name                   = "my-deployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = var.deployment_template_id

  observability = {
    deployment_id = ec_deployment.monitoring.id
    ref_id        = ec_deployment.monitoring.elasticsearch.ref_id
  }

  elasticsearch = {
    hot = {
      autoscaling = {}
    }
  }

  kibana = {}
}

provider "elasticstack" {
  # Use our Elastic Cloud deployment outputs for connection details.
  # This also allows the provider to create the proper relationships between the two resources.
  elasticsearch {
    endpoints = ["${ec_deployment.cluster.elasticsearch.https_endpoint}"]
    username  = ec_deployment.cluster.elasticsearch_username
    password  = ec_deployment.cluster.elasticsearch_password
  }

  kibana {
    endpoints = ["${ec_deployment.cluster.kibana.https_endpoint}"]
  }
}

provider "elasticstack" {
  # Use our Elastic Cloud deployment outputs for connection details.
  # This also allows the provider to create the proper relationships between the two resources.
  elasticsearch {
    endpoints = ["${ec_deployment.monitoring.elasticsearch.https_endpoint}"]
    username  = ec_deployment.monitoring.elasticsearch_username
    password  = ec_deployment.monitoring.elasticsearch_password
  }
  alias = "monitoring"
}

#	resource "elasticstack_kibana_action_connector" "test" {
#	  name         = "test connector 1"
#	  config = jsonencode({
#		createIncidentJson = "{}"
#		createIncidentResponseKey = "key"
#		createIncidentUrl = "https://www.elastic.co/"
#		getIncidentResponseExternalTitleKey = "title"
#		getIncidentUrl = "https://www.elastic.co/"
#		updateIncidentJson = "{}"
#		updateIncidentUrl = "https://elasticsearch.com/"
#		viewIncidentUrl = "https://www.elastic.co/"
#		createIncidentMethod = "put"
#	  })
#	  secrets = jsonencode({
#		user = "user2"
#		password = "password2"
#	  })
#	  connector_type_id = ".cases-webhook"
#	}
#
#  	resource "elasticstack_kibana_action_connector" "test2" {
#	  name         = "test connector 2"
#    connector_id = "1d30b67b-f90b-4e28-87c2-137cba361509"
#	  config = jsonencode({
#		createIncidentJson = "{}"
#		createIncidentResponseKey = "key"
#		createIncidentUrl = "https://www.elastic.co/"
#		getIncidentResponseExternalTitleKey = "title"
#		getIncidentUrl = "https://www.elastic.co/"
#		updateIncidentJson = "{}"
#		updateIncidentUrl = "https://elasticsearch.com/"
#		viewIncidentUrl = "https://www.elastic.co/"
#		createIncidentMethod = "put"
#	  })
#	  secrets = jsonencode({
#		user = "user2"
#		password = "password2"
#	  })
#	  connector_type_id = ".cases-webhook"
#	}
#
#resource "elasticstack_elasticsearch_security_api_key" "cross_cluster_key" {
#  name = "My Cross-Cluster API Key"
#  type = "cross_cluster"
#  # Define access permissions for cross-cluster operations
#  access = {
#    # Grant replication access to specific indices
#    replication = [
#      {
#        names = ["archive-test-1-*"]
#      }
#    ]
#
#    search = [
#      {
#        names = ["log-1-*", "metrics-1-*"]
#      }
#    ]
#  }
#  # Set the expiration for the API key
#  expiration = "30d"
#  # Set arbitrary metadata
#  metadata = jsonencode({
#    description = "Cross-cluster key for production environment"
#    environment = "production"
#    team        = "platform"
#  })
#}
#output "cross_cluster_api_key" {
#  value     = elasticstack_elasticsearch_security_api_key.cross_cluster_key
#  sensitive = true
#}

#resource "elasticstack_kibana_action_connector" "my-new-connector7" {
#  name              = "human-uuid"
##  connector_id      = "lugoz-safes-rusin-bubov-fytex-cydeb"
# connector_id      = "abc69090-6342-4f3f-b236-a3fd48635227"
#  connector_type_id = ".slack_api"
#  secrets = jsonencode({
#    token = "dummy value"
#  })
#}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "Test Detection Rule (updated 17)"
  description = "A test detection rule"
  type        = "query"
  severity    = "medium"
  risk_score  = 50
  enabled     = true
  query       = "user.name:*"
  language    = "kuery"

  tags = ["test", "terraform"]
}

#resource "elasticstack_kibana_security_detection_rule" "test" {
#  name         = "Test Threat Match Rule"
#  type         = "threat_match"
#  query        = "destination.ip:*"
#  language     = "kuery"
#  enabled      = true
#  description  = "Test threat match security detection rule"
#  severity     = "high"
#  risk_score   = 80
#  from         = "now-6m"
#  to           = "now"
#  interval     = "5m"
#  index        = ["logs-*"]
#  threat_index = ["threat-intel-*"]
#  threat_query = "threat.indicator.type:ip"
#  
#  threat_mapping = [
#    {
#      entries = [
#        {
#          field = "destination.ip"
#          type  = "mapping"
#          value = "threat.indicator.ip"
#        }
#      ]
#    }
#  ]
#}

#resource "elasticstack_kibana_security_detection_rule" "test" {
#  name        = "Test Threshold Rule"
#  type        = "threshold"
#  query       = "event.action:login"
#  language    = "kuery"
#  enabled     = true
#  description = "Test threshold security detection rule"
#  severity    = "medium"
#  risk_score  = 55
#  from        = "now-6m"
#  to          = "now"
#  interval    = "5m"
#  index       = ["logs-*"]
#
#  threshold = {
#    value = 10
#    field = ["user.name"]
#    #cardinality = [ # TODO test without cardinality
#    #  {
#    #    field = "source.ip"
#    #    value = 5
#    #  }
#    #]
#  }
#}

#resource "elasticstack_kibana_security_detection_rule" "example" {
#  name        = "Suspicious Process Activity"
#  description = "Detects suspicious process execution patterns"
#  type        = "query"
#  query       = "process.name : (cmd.exe or powershell.exe) and user.name : admin*"
#  language    = "kuery"
#  severity    = "high"
#  risk        = 75
#  enabled     = true
#  
#  tags        = ["security", "windows", "process"]
#  interval    = "5m"
#  from        = "now-6m"
#  to          = "now"
#  
#  author      = ["Security Team"]
#  references  = ["https://attack.mitre.org/techniques/T1059/"]
#}

#resource "elasticstack_kibana_security_detection_rule" "test" {
#  name         = "TEST "
#  type         = "threat_match"
#  query        = "destination.ip:*"
#  language     = "kuery"
#  enabled      = true
#  description  = "Test threat match security detection rule"
#  severity     = "high"
#  risk_score   = 80
#  from         = "now-6m"
#  to           = "now"
#  interval     = "5m"
#  index        = ["logs-*"]
#  threat_index = ["threat-intel-*"]
#  threat_query = "threat.indicator.type:ip"
#
#  threat_mapping = [
#    {
#      entries = [
#        {
#          field = "destination.ip"
#          type  = "mapping"
#          value = "threat.indicator.ip"
#        }
#      ]
#    },
#    {
#      entries = [
#        {
#          field = "source.ip"
#          type  = "mapping"
#          value = "threat.indicator.ip"
#        }
#      ]
#    }
#  ]
#}
```

Notice that the Elastic Stack  provider setup with empty `elasticsearch {}` block, since we'll be using an `elasticsearch_connection` block
for each of our resources, to point to the Elastic Cloud deployment.



## Managing stack resources and configuration

Now we can add resources to these deployments as follows:

```terraform
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
}

# Configuring my cluster with an index template as well.
resource "elasticstack_elasticsearch_index_template" "my_template" {
  name = "my_ingest_1"

  priority       = 42
  index_patterns = ["server-logs*"]

  template {
    alias {
      name = "my_template_test"
    }

    settings = jsonencode({
      number_of_shards = "3"
    })

    mappings = jsonencode({
      properties : {
        "@timestamp" : { "type" : "date" },
        "username" : { "type" : "keyword" }
      }
    })
  }
}

# Defining a user for viewing monitoring
resource "elasticstack_elasticsearch_security_user" "monitoring_user" {
  # Explicitly select the monitoring provider here
  provider = elasticstack.monitoring

  username = "monitoring_viewer"

  # Password is cleartext here for comfort, but there's also a hashed password option
  password = "mysecretpassword"
  roles    = ["reader"]

  # Set the custom metadata for this user
  metadata = jsonencode({
    "env"    = "testing"
    "open"   = false
    "number" = 49
  })
}
```

Note that resources can be targed to certain deployments using the `provider` attribute. 
