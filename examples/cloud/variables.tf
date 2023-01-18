#This needs to be set in order to run the example. Best to use the TF_VAR_ec_apikey environment variable.
variable "ec_apikey" {
  type    = string
  default = ""
}

variable "region" {
  type    = string
  default = "gcp-us-central1"
}

variable "deployment_template_id" {
  type    = string
  default = "gcp-storage-optimized"
}
