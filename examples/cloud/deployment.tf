# Creating deployments on Elastic Cloud GCP region with elasticsearch and kibana components. One deployment is a dedicated monitor for the other. 

resource "ec_deployment" "monitoring" {
  region                 = var.region
  name                   = "my-monitoring-deployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = var.deployment_template_id

  elasticsearch {}
  kibana {}
}

resource "ec_deployment" "cluster" {
  region                 = var.region
  name                   = "my-deployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = var.deployment_template_id

  observability {
    deployment_id = ec_deployment.monitoring.id
    ref_id        = ec_deployment.monitoring.elasticsearch[0].ref_id
  }

  elasticsearch {}
  kibana {}
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = var.region
}

provider "elasticstack" {
  # Use our Elastic Cloud deployment outputs for connection details.
  # This also allows the provider to create the proper relationships between the two resources.
  elasticsearch {
    endpoints = ["${ec_deployment.cluster.elasticsearch[0].https_endpoint}"]
    username  = ec_deployment.cluster.elasticsearch_username
    password  = ec_deployment.cluster.elasticsearch_password
  }
}

provider "elasticstack" {
  # Use our Elastic Cloud deployment outputs for connection details.
  # This also allows the provider to create the proper relationships between the two resources.
  elasticsearch {
    endpoints = ["${ec_deployment.monitoring.elasticsearch[0].https_endpoint}"]
    username  = ec_deployment.monitoring.elasticsearch_username
    password  = ec_deployment.monitoring.elasticsearch_password
  }
  alias = "monitoring"
}

