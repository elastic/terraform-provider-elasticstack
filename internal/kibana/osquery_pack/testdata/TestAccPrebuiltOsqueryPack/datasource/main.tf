variable "prebuilt_pack_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_kibana_osquery_pack" "test" {
  pack_id = var.prebuilt_pack_id
}
