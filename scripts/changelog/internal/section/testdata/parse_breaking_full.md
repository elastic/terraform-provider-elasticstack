## Changelog

Customer impact: breaking
Summary: Remove deprecated attribute from elasticstack_kibana_slo

### Breaking changes

The `legacy_mode` attribute has been removed.

- Attribute `legacy_mode` is no longer accepted.
- Existing configs referencing `legacy_mode` must be updated.

```hcl
# Before
resource "elasticstack_kibana_slo" "example" {
  legacy_mode = true
}

# After — remove the attribute entirely
resource "elasticstack_kibana_slo" "example" {
}
```

## Other section

Not part of breaking changes.
