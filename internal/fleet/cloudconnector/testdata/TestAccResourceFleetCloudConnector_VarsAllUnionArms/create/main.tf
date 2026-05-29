variable "name" {
  type = string
}

variable "gcp_credentials_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_cloud_connector" "test" {
  name           = var.name
  cloud_provider = "gcp"
  vars = {
    service_account = {
      type  = "text"
      value = "my-sa@project.iam.gserviceaccount.com"
    }
    audience = {
      type  = "text"
      value = "//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/pool/providers/provider"
    }
    gcp_credentials_cloud_connector_id = {
      type  = "text"
      value = var.gcp_credentials_id
    }
    custom_string = {
      string = "bare-string-arm"
    }
    custom_number = {
      number = 42.5
    }
    custom_bool = {
      bool = true
    }
    custom_struct_text = {
      type  = "text"
      value = "structured-text-arm"
    }
  }
}
