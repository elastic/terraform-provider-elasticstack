variable "policy_name" {
  type        = string
  description = "Name of the ILM policy"
}

variable "fleet_system_version" {
  type        = string
  description = "System package version for Fleet integration (must match version installed in test PreCheck)"
}
