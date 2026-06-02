Manages the deployment lifecycle (start, scale, stop) of an existing Elasticsearch ML trained model.

This resource does not upload or create the underlying trained model; it only manages the deployment state.
On Terraform destroy the resource stops (undeploys) the model deployment.
