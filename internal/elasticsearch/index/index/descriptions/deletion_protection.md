Whether to allow Terraform to destroy the index. Unless this field is set to `false` in Terraform state, a `terraform destroy` or `terraform apply` command that deletes the index will fail.

Destroying (or replacing) a protected index is a **two-step process**: the check always runs against the index's last-applied state, not the new plan, so you cannot set `deletion_protection = false` in the same `terraform apply` that also destroys or replaces the index (for example, a configuration change that forces replacement).

1. Apply a change that sets `deletion_protection = false` on its own.
2. Only then run the `terraform apply` or `terraform destroy` that removes or replaces the index.
