variable "policy_name" {
  type        = string
  description = "Name for the agent policy"
}

variable "skip_destroy" {
  type        = bool
  description = "Whether to skip destruction of the policy"
  default     = false
}

