variable "unknown_pack_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_kibana_osquery_pack" "missing" {
  pack_id = var.unknown_pack_id
}
