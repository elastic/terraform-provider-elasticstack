variable "name" {
  type        = string
  description = "The name of the SLO."
}

variable "tags" {
  type        = list(string)
  description = "A list of tags to associate with the SLO."
  default     = ["test", "terraform"]
}
