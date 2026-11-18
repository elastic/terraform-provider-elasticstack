# Shared configuration for the write-only secret acceptance tests. Every
# optional input defaults to "" which maps to null, so a single directory can
# express each credential combination by varying ConfigVariables. The literal
# "<empty>" maps to an empty string to exercise empty-value validation.
provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

variable "label" {
  type    = string
  default = "a"
}

variable "certificate" {
  type    = string
  default = ""
}

variable "key" {
  type      = string
  default   = ""
  sensitive = true
}

variable "key_wo" {
  type      = string
  default   = ""
  sensitive = true
}

variable "username" {
  type    = string
  default = ""
}

variable "password" {
  type      = string
  default   = ""
  sensitive = true
}

variable "password_wo" {
  type      = string
  default   = ""
  sensitive = true
}

variable "api_key" {
  type      = string
  default   = ""
  sensitive = true
}

variable "api_key_wo" {
  type      = string
  default   = ""
  sensitive = true
}

variable "headers" {
  type    = bool
  default = false
}

locals {
  certificate = var.certificate == "" ? null : var.certificate
  key         = var.key == "" ? null : (var.key == "<empty>" ? "" : var.key)
  key_wo      = var.key_wo == "" ? null : (var.key_wo == "<empty>" ? "" : var.key_wo)
  username    = var.username == "" ? null : var.username
  password    = var.password == "" ? null : var.password
  password_wo = var.password_wo == "" ? null : var.password_wo
  api_key     = var.api_key == "" ? null : var.api_key
  api_key_wo  = var.api_key_wo == "" ? null : var.api_key_wo

  ssl_set  = local.certificate != null || local.key != null || local.key_wo != null
  auth_set = var.headers || local.username != null || local.password != null || local.password_wo != null || local.api_key != null || local.api_key_wo != null
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "WO Agent Download Source ${var.suffix} ${var.label}"
  source_id = "agent-download-source-wo-${var.suffix}"
  default   = false
  host      = "https://artifacts.example.com/elastic-agent-wo"
  space_ids = ["default"]

  ssl = local.ssl_set ? {
    certificate = local.certificate
    key         = local.key
    key_wo      = local.key_wo
  } : null

  auth = local.auth_set ? {
    username    = local.username
    password    = local.password
    password_wo = local.password_wo
    api_key     = local.api_key
    api_key_wo  = local.api_key_wo
    headers     = var.headers ? [{ key = "X-A", value = "1" }] : null
  } : null
}
