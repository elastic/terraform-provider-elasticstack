Opt-in flag for **create-time** adoption of an index that already exists when Terraform runs create (for example after a replacement race or when managing an index created out-of-band).

When `false` or unset (the default), the resource behaves as before and create always attempts a new index.

After a resource is already in state, this attribute has **no create-time effect**: toggling it only changes configuration; apply does not re-run create, so it is effectively a planning no-op for lifecycle behavior.

The adoption path runs only for **static** index `name` values. For date math index names, the provider emits a warning that `use_existing` does not apply and proceeds with the normal create flow.

When adoption runs, every **static** index setting you **explicitly** set in configuration must match the existing index; any mismatch returns an error and the provider does not change the cluster.

After a successful adopt, Terraform fully manages the index: subsequent reads, updates, and destroys follow the normal resource behavior.
