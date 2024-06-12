//monitoring cluster
output "monitoring_kibana_https_endpoint" {
  value = ec_deployment.monitoring.kibana.https_endpoint
}

output "monitoring_elasticsearch_https_endpoint" {
  value = ec_deployment.monitoring.elasticsearch.https_endpoint
}

output "monitoring_elasticsearch_username" {
  value = ec_deployment.monitoring.elasticsearch_username
}

output "monitoring_elasticsearch_password" {
  value = nonsensitive(ec_deployment.monitoring.elasticsearch_password)
}

output "monitoring_elasticsearch_id" {
  value = ec_deployment.monitoring.elasticsearch.resource_id
}

output "monitoring_deployment_id" {
  value = ec_deployment.monitoring.id
}

//data cluster
output "data_kibana_https_endpoint" {
  value = ec_deployment.cluster.kibana.https_endpoint
}

output "data_elasticsearch_https_endpoint" {
  value = ec_deployment.cluster.elasticsearch.https_endpoint
}

output "data_elasticsearch_username" {
  value = ec_deployment.cluster.elasticsearch_username
}

output "data_elasticsearch_password" {
  value = nonsensitive(ec_deployment.cluster.elasticsearch_password)
}

output "data_elasticsearch_id" {
  value = ec_deployment.cluster.elasticsearch.resource_id
}

output "data_deployment_id" {
  value = ec_deployment.cluster.id
}
